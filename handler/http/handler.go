package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/model"
	errors "github.com/tomogoma/go-typed-errors"
)

type contextKey string

type Auth interface {
	errors.AllErrChecker

	RegisterFirst(loginType, userType, id string, secret []byte) (*model.User, error)
	CanRegisterFirst() (bool, error)
	RegisterSelf(loginType, userType, id string, secret []byte) (*model.User, error)
	RegisterSelfByLockedPhone(userType, devID, number string, password []byte) (*model.User, error)
	RegisterOther(JWT, newLoginType, userType, id, groupID string) (*model.User, error)

	UpdateIdentifier(JWT, forUserID, loginType, newId string) (*model.User, error)

	UpdatePassword(JWT string, old, newPass []byte) error
	SetPassword(loginType, onAddr string, dbt, pass []byte) (*model.VerifLogin, error)

	SendVerCode(JWT, loginType, toAddr string) (*model.DBTStatus, error)
	SendPassResetCode(loginType, toAddr string) (*model.DBTStatus, error)

	VerifyAndExtendDBT(lt, forAddr string, dbt []byte) (string, error)
	VerifyDBT(loginType, forAddr string, dbt []byte) (*model.VerifLogin, error)

	Login(loginType, identifier string, password []byte) (*model.User, error)

	Users(JWT string, q model.UsersQuery, offset, count string) ([]model.User, error)
	GetUserDetails(JWT, userID string) (*model.User, error)

	Groups(JWT, offset, count string) ([]model.Group, error)
}

type Guard interface {
	APIKeyValid(key string) (string, error)
}

type handler struct {
	errors.NotImplErrCheck
	errors.AuthErrCheck
	errors.ClErrCheck
	errors.NotFoundErrCheck

	auth   Auth
	guard  Guard
	logger logging.Logger
}

const (
	internalErrorMessage = "whoops! Something wicked happened"

	keyLoginType    = "loginType"
	keySelfReg      = "selfReg"
	keyAPIKey       = "x-api-key"
	keyToken        = "token"
	keyOTP          = "OTP"
	keyExtend       = "extend"
	keyOffset       = "offset"
	keyCount        = "count"
	keyUserID       = "userID"
	keyGroupID      = "groupID"
	keyAcl          = "acl"
	keyGroup        = "group"
	keyMatchAllACLs = "matchAllACLs"
	keyMatchAll     = "matchAll"

	ctxKeyLog = contextKey("log")

	valTrue   = "true"
	valDevice = "device"
)

func NewHandler(a Auth, g Guard, l logging.Logger, allowedOrigins []string) (http.Handler, error) {
	if a == nil {
		return nil, errors.New("Auth was nil")
	}
	if g == nil {
		return nil, errors.New("Guard was nil")
	}
	if l == nil {
		return nil, errors.New("Logger was nil")
	}

	r := mux.NewRouter().PathPrefix(config.WebRootURL()).Subrouter()
	handler{auth: a, guard: g, logger: l}.handleRoute(r)

	headersOk := handlers.AllowedHeaders([]string{
		"X-Requested-With", "Accept", "Content-Type", "Content-Length",
		"Accept-Encoding", "X-CSRF-Token", "Authorization", "X-api-key",
	})
	originsOk := handlers.AllowedOrigins(allowedOrigins)
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	return handlers.CORS(headersOk, originsOk, methodsOk)(r), nil
}

func (s handler) handleRoute(r *mux.Router) {

	r.PathPrefix("/status").
		Methods(http.MethodGet).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleStatus)))

	r.PathPrefix("/first_user").
		Methods(http.MethodPut).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleRegisterFirst)))

	r.PathPrefix("/reset_password/send_otp").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleSendPassResetCode)))

	r.PathPrefix("/reset_password").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleResetPass)))

	r.PathPrefix("/groups").
		Methods(http.MethodGet).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleGroups)))

	r.PathPrefix("/users/{" + keyUserID + "}/set_group/{" + keyGroupID + "}").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleSetUserGroup)))

	r.PathPrefix("/users/{" + keyUserID + "}").
		Methods(http.MethodGet).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleUserDetails)))

	r.PathPrefix("/users/{" + keyUserID + "}").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleUpdate)))

	r.PathPrefix("/users").
		Methods(http.MethodGet).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleUsers)))

	r.PathPrefix("/{" + keyLoginType + "}/register").
		Methods(http.MethodPut).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleRegistration)))

	r.PathPrefix("/{" + keyLoginType + "}/verify/{" + keyOTP + "}").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleVerifyCode)))

	r.PathPrefix("/{" + keyLoginType + "}/verify").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleSendVerifCode)))

	r.PathPrefix("/{" + keyLoginType + "}/login").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleLogin)))

	r.PathPrefix("/" + config.DocsPath).
		Handler(http.FileServer(http.Dir(config.DefaultDocsDir())))

	r.NotFoundHandler = http.HandlerFunc(s.prepLogger(s.notFoundHandler))
}

func (s handler) prepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log := s.logger.WithHTTPRequest(r).
			WithField(logging.FieldTransID, uuid.New())

		log.WithFields(map[string]interface{}{
			logging.FieldURL:        r.URL.Path,
			logging.FieldHTTPMethod: r.Method,
		}).Info("new request")

		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *handler) guardRoute(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		APIKey := r.Header.Get(keyAPIKey)
		clUsrID, err := s.guard.APIKeyValid(APIKey)
		log := r.Context().Value(ctxKeyLog).(logging.Logger).
			WithField(logging.FieldClientAppUserID, clUsrID)
		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		if err != nil {
			s.handleError(w, r.WithContext(ctx), nil, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// unmarshalJSONOrRespondError returns true if json is extracted from
// data into req successfully, otherwise, it writes an error response into
// w and returns false.
// The Context in r should contain a logging.Logger with key ctxKeyLog
// for logging in case of error
func (s *handler) unmarshalJSONOrRespondError(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	dataB, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err = errors.NewClientf("Failed to read request body: %v", err)
		s.handleError(w, r, nil, err)
		return false
	}
	if err = json.Unmarshal(dataB, req); err != nil {
		err = errors.NewClientf("failed to unmarshal JSON request from body: %v", err)
		s.handleError(w, r, nil, err)
		return false
	}
	return true
}

/**
 * @api {get} /status Status
 * @apiName Status
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiSuccess {String} name Micro-service name.
 * @apiSuccess  {String} version Current running version.
 * @apiSuccess {String} description Short description of the micro-service.
 * @apiSuccess {String} canonicalName Canonical name of the micro-service.
 * @apiSuccess {String} needRegSuper true if a super-user has been registered, false otherwise.
 *
 */
func (s *handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	canRegFrst, err := s.auth.CanRegisterFirst()
	s.respondOn(w, r, nil, struct {
		Name          string `json:"name"`
		Version       string `json:"version"`
		Description   string `json:"description"`
		CanonicalName string `json:"canonicalName"`
		NeedRegSuper  bool   `json:"needRegSuper"`
	}{
		Name:          config.Name,
		Version:       config.VersionFull,
		Description:   config.Description,
		CanonicalName: config.CanonicalWebName(),
		NeedRegSuper:  canRegFrst,
	}, http.StatusOK, err)
}

/**
 * @api {get} /users Get Users
 * @apiName GetUsers
 * @apiVersion 0.1.1
 * @apiGroup Auth
 * @apiPermission ^admin
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (URL Query Parameters) {String} token the JWT accessed during auth.
 * @apiParam (URL Query Parameters) {Number} [offset=0] The beginning index to fetch groups.
 * @apiParam (URL Query Parameters) {Number} [count=10] The maximum number of groups to fetch.
 * @apiParam (URL Query Parameters) {String} [group] Filter by group name.
	one can have multiple groups e.g. ?group=admin&group=staff,
	multiple group names are always filtered using the OR operator.
 * @apiParam (URL Query Parameters) {String{0-10}=gt_[number],lt_[number],[number],gteq_[number],lteq_[number]} [acl] Filter by access levels:
 * - gt_[number] - access level greater than number e.g. gt_5
 * - lt_[number] - access level less than number e.g. lt_5
 * - [number] - access level equal to number e.g. 5
 * - gteq_[number] - access level greater than or equal to number e.g. gteq_5
 * - lteq_[number] - access level less than or equal to  number e.g. lteq_5
 * - one can have multiple filters e.g. ?acl=gt_5&acl=lteq_9&matchAllACLs=true to get acl in (5 < acl <= 9)
 * @apiParam (URL Query Parameters) {String=true,false} [matchAllACLs=false] Setting
	this to true will force all acl's provided to be matched using
	the AND operator, otherwise uses the OR operator.
 * @apiParam (URL Query Parameters) {String=true,false} [matchAll=false]
	Setting this to true will force acl,group filters to be matched using the AND
	operator, otherwise uses the OR operator.
 * @apiParam (URL Query Parameters) {String} token The JWT provided during auth.
 *
 * @apiSuccess {Object[]} json-body JSON array of <a href="#api-Objects-Group">groups</a>
 *
 */
func (s *handler) handleUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := struct {
		JWT          string   `json:"token"`
		Offset       string   `json:"offset"`
		Count        string   `json:"count"`
		Groups       []string `json:"group"`
		ACLs         []string `json:"acl"`
		MatchAllACLs string   `json:"matchAllACLs"`
		MatchAll     string   `json:"matchAll"`
	}{
		JWT:          q.Get(keyToken),
		Offset:       q.Get(keyOffset),
		Count:        q.Get(keyCount),
		MatchAllACLs: q.Get(keyMatchAllACLs),
		MatchAll:     q.Get(keyMatchAll),
		Groups:       q[keyGroup],
		ACLs:         q[keyAcl],
	}
	uq := model.UsersQuery{
		AccessLevelsIn: req.ACLs,
		MatchAllACLs:   strings.EqualFold(req.MatchAllACLs, valTrue),
		GroupNamesIn:   req.Groups,
		MatchAll:       strings.EqualFold(req.MatchAll, valTrue),
	}
	usrs, err := s.auth.Users(req.JWT, uq, req.Offset, req.Count)
	s.respondOn(w, r, req, NewUsers(usrs), http.StatusOK, err)
}

/**
 * @api {get} /users/:userID User Details
 * @apiName UserDetails
 * @apiVersion 0.1.1
 * @apiGroup Auth
 * @apiPermission owner|^staff
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (URL Parameters) {String} :userID The ID of the
	<a href="#api-Objects-User">User</a> whose details are sort.
 *
 * @apiParam (URL Query Parameters) {String} token The JWT provided during auth.
 *
 * @apiUse User
 *
 */
func (s *handler) handleUserDetails(w http.ResponseWriter, r *http.Request) {
	req := struct {
		UserID string `json:"userID"`
		JWT    string `json:"token"`
	}{
		UserID: mux.Vars(r)[keyUserID],
		JWT:    r.URL.Query().Get(keyToken),
	}
	usr, err := s.auth.GetUserDetails(req.JWT, req.UserID)
	s.respondOn(w, r, req, NewUser(usr), http.StatusOK, err)
}

/**
 * @api {get} /groups Get Groups
 * @apiName GetGroups
 * @apiVersion 0.1.1
 * @apiGroup Auth
 * @apiPermission ^admin
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (URL Query Parameters) {String} token the JWT accessed during auth.
 * @apiParam (URL Query Parameters) {Number} [offset=0] The beginning index to fetch groups.
 * @apiParam (URL Query Parameters) {Number} [count=10] The maximum number of groups to fetch.
 *
 * @apiParam (URL Query Parameters) {String} token The JWT provided during auth.
 *
 * @apiSuccess {Object[]} json-body JSON array of <a href="#api-Objects-Group">groups</a>
 *
 */
func (s *handler) handleGroups(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := struct {
		JWT    string `json:"token"`
		Offset string `json:"offset"`
		Count  string `json:"count"`
	}{
		JWT:    q.Get(keyToken),
		Offset: q.Get(keyOffset),
		Count:  q.Get(keyCount),
	}
	grps, err := s.auth.Groups(req.JWT, req.Offset, req.Count)
	s.respondOn(w, r, req, NewGroups(grps), http.StatusOK, err)
}

/**
 * @api {put} /first_user First User
 * @apiDescription Register the first super-user (super admin)
 * @apiPermission anyone
 * @apiName FirstUser
 * @apiVersion 0.1.0
 * @apiGroup Setup
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (JSON Request Body) {String=individual,company} userType Type of user.
 * @apiParam (JSON Request Body) {String=usernames,phones,emails,facebook} loginType Type of identifier.
 * @apiParam (JSON Request Body) {String} identifier The user's unique loginType identifier.
 * @apiParam (JSON Request Body) {String} secret The users password
 *
 * @apiSuccess (Success 201) {Object} json-body See <a href="#api-Objects-User">User</a> for details.
 *
 */
func (s *handler) handleRegisterFirst(w http.ResponseWriter, r *http.Request) {
	req := &struct {
		UserType   string `json:"userType"`
		LT         string `json:"loginType"`
		Identifier string `json:"identifier"`
		Secret     string `json:"secret"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, req) {
		return
	}
	usr, err := s.auth.RegisterFirst(req.LT, req.UserType, req.Identifier, []byte(req.Secret))
	req.Secret = "" // prevent logging passwords.
	s.respondOn(w, r, req, NewUser(usr), http.StatusCreated, err)
}

/**
 * @api {put} /:loginType/register Register
 * @apiDescription  Register new user.
 * @apiPermission ^admin for registering other
 * @apiName Register
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (URL Parameters) {String=usernames,emails,phones,facebook} loginType type of identifier in JSON Body
 *
 * @apiParam (URL Query Parameters) {String=true,device,} selfReg Whether registering self or not:
 * - true for self registration
 * - device for self registration by unique device ID
 * - not-provided for admin to register any other user
 *
 * @apiParam (JSON Request Body) {String=individual,company} userType Type of user.
 * @apiParam (JSON Request Body) {String} identifier The 'username' corresponding to loginType.
 * @apiParam (JSON Request Body) {String} [secret] The user's password - required when selfReg not set.
 * @apiParam (JSON Request Body) {String} [groupID] groupID to add this user to - required when selfReg not set.
 * @apiParam (JSON Request Body) {String} [deviceID] the unique device ID for the user - required when selfReg=device.
 *
 * @apiSuccess (Success 201) {Object} json-body See <a href="#api-Objects-User">User</a> for details.
 *
 */
func (s *handler) handleRegistration(w http.ResponseWriter, r *http.Request) {
	req := &struct {
		UserType   string `json:"userType"`
		Identifier string `json:"identifier"`
		Secret     string `json:"secret"`
		GroupID    string `json:"groupID"`
		DevID      string `json:"deviceID"`
		LT         string `json:"loginType"`
		SelfReg    string `json:"selfReg"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, req) {
		return
	}
	vars := mux.Vars(r)
	req.LT = vars[keyLoginType]
	req.SelfReg = r.URL.Query().Get(keySelfReg)
	var usr *model.User
	var err error
	switch strings.ToLower(req.SelfReg) {
	case valTrue:
		usr, err = s.auth.RegisterSelf(req.LT, req.UserType, req.Identifier, []byte(req.Secret))
	case valDevice:
		usr, err = s.auth.RegisterSelfByLockedPhone(req.UserType, req.DevID, req.Identifier, []byte(req.Secret))
	default:
		JWT := r.URL.Query().Get(keyToken)
		usr, err = s.auth.RegisterOther(JWT, req.LT, req.UserType, req.Identifier, req.GroupID)
	}
	req.Secret = "" // prevent logging passwords.
	s.respondOn(w, r, req, NewUser(usr), http.StatusCreated, err)
}

/**
 * @api {POST} /:loginType/login Login
 * @apiDescription User login.
 * See <a href="#api-Auth-Register">Register</a> for loginType options.
 * @apiName Login
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 * @apiHeader Authorization Basic auth containing loginType's identifier and password in the format
	'Basic: base64Of(identifier:password)'
 *
 * @apiParam (URL Parameters) {String=usernames,emails,phones,facebook} loginType type of identifier in Authorization header.
 *
 * @apiParam (URL Parameters) {String=usernames,emails,phones,facebook} loginType type of identifier in JSON Body
 *
 * @apiUse User
 *
 */
func (s *handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	req := struct {
		LT         string `json:"loginType"`
		Identifier string `json:"identifier"`
		Found      bool   `json:"isBasicAuthFound"`
	}{}
	vars := mux.Vars(r)
	req.LT = vars[keyLoginType]
	var secret string // separate from req to prevent logging passwords.
	req.Identifier, secret, req.Found = r.BasicAuth()
	if !req.Found {
		err := errors.NewUnauthorized("missing BasicAuth details")
		s.handleError(w, r, req, err)
		return
	}
	usr, err := s.auth.Login(req.LT, req.Identifier, []byte(secret))
	s.respondOn(w, r, req, NewUser(usr), http.StatusOK, err)
}

/**
 * @api {POST} /users/:userID Update Identifier
 * @apiDescription Update (or set for first time) the identifier details for loginType.
 * @apiName UpdateIdentifier
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (URL Parameters) {String} userID The ID of the <a href="#api-Objects-User">user</a> to update.
 *
 * @apiParam (URL Query Parameters) {String} token the JWT provided during login.
 *
 * @apiParam (JSON Request Body) {String=usernames,emails,phones,facebook} loginType type of identifier in JSON Body
 * @apiParam (JSON Request Body) {String} identifier The new loginType's unique identifier.
 *
 * @apiUse User
 *
 */
func (s *handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	req := &struct {
		Identifier string `json:"identifier"`
		UserID     string `json:"userID"`
		LT         string `json:"loginType"`
		JWT        string `json:"token"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, req) {
		return
	}
	req.JWT = r.URL.Query().Get(keyToken)
	req.UserID = mux.Vars(r)[keyUserID]
	usr, err := s.auth.UpdateIdentifier(req.JWT, req.UserID, req.LT, req.Identifier)
	s.respondOn(w, r, req, NewUser(usr), http.StatusOK, err)
}

/**
 * @api {POST} /users/:userID/set_group/:groupID Set User's Group
 * @apiDescription Assign group to user.
 * @apiName SetUserGroup
 * @apiVersion 0.1.0
 * @apiPermission ^admin
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (URL Parameters) {String} userID The ID of the <a href="#api-Objects-User">user</a> to update.
 * @apiParam (URL Parameters) {String} groupID The ID of the <a href="#api-Objects-Group">group</a> to assign the user to.
 *
 * @apiParam (URL Query Parameters) {String} token the JWT provided during login.
 *
 * @apiUse User
 *
 */
func (s *handler) handleSetUserGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &struct {
		UserID  string `json:"userID"`
		GroupID string `json:"groupID"`
		JWT     string `json:"token"`
	}{
		UserID:  vars[keyUserID],
		GroupID: vars[keyGroupID],
		JWT:     r.URL.Query().Get(keyToken),
	}
	s.respondOn(w, r, req, nil, http.StatusOK, errors.NewNotImplemented())
}

/**
 * @api {POST} /:loginType/verify Send Verification Code
 * @apiDescription Send OTP to identifier of type loginType for purpose of verifying identifier.
 * @apiName SendVerificationCode
 * @apiVersion 0.2.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (URL Parameters) {String=usernames,emails,phones,facebook} loginType type of identifier in JSON Body
 *
 * @apiParam (URL Query Parameters) {String} token the JWT provided during login.
 *
 * @apiParam (JSON Request Body) {String} identifier The loginType's address to be verified.
 *
 * @apiUse OTPStatus
 *
 */
func (s *handler) handleSendVerifCode(w http.ResponseWriter, r *http.Request) {
	req := &struct {
		LT     string `json:"loginType"`
		ToAddr string `json:"identifier"`
		JWT    string `json:"token"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, req) {
		return
	}
	vars := mux.Vars(r)
	req.LT = vars[keyLoginType]
	req.JWT = r.URL.Query().Get(keyToken)
	dbtStatus, err := s.auth.SendVerCode(req.JWT, req.LT, req.ToAddr)
	s.respondOn(w, r, req, NewDBTStatus(dbtStatus), http.StatusOK, err)
}

/**
 * @api {POST} /:loginType/verify/:OTP Verify OTP
 * @apiDescription Verify OTP.
 * See <a href="#api-Auth-Register">Register</a> for loginType options.
 * extend can be set to "true" if intent on extending the expiry of the OTP.
 * @apiName VerifyOTP
 * @apiVersion 0.2.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (URL Parameters) {String=usernames,emails,phones,facebook} loginType type of identifier to verify.
 *
 * @apiParam (URL Query Parameters) {String=true,} extend set true to return an extended expiry period OTP.
 *
 * @apiParam (JSON Request Body) {String} identifier The loginType's address to whom the OTP was sent.
 *
 * @apiSuccess {String} [OTP] (if extending OTP) the new OTP with extended expiry
 *
 * @apiUseSuccess {Object} json-body (if not extending OTP) The verified <a href="#api-Objects-VerifLogin">loginType identifier</a>.
 *
 */
func (s *handler) handleVerifyCode(w http.ResponseWriter, r *http.Request) {
	req := &struct {
		Identifier string `json:"identifier"`
		LT         string `json:"loginType"`
		DBT        string `json:"OTP"`
		Extend     string `json:"extend"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, req) {
		return
	}
	vars := mux.Vars(r)
	req.LT = vars[keyLoginType]
	req.DBT = vars[keyOTP]
	q := r.URL.Query()
	req.Extend = q.Get(keyExtend)

	var resp interface{}
	var err error
	if strings.EqualFold(req.Extend, valTrue) {
		var dbt string
		dbt, err = s.auth.VerifyAndExtendDBT(req.LT, req.Identifier, []byte(req.DBT))
		resp = struct {
			OTP string `json:"OTP"`
		}{OTP: dbt}
	} else {
		var vl *model.VerifLogin
		vl, err = s.auth.VerifyDBT(req.LT, req.Identifier, []byte(req.DBT))
		resp = NewVerifLogin(vl)
	}
	s.respondOn(w, r, req, resp, http.StatusOK, err)
}

/**
 * @api {POST} /reset_password/send_otp Send Password Reset OTP
 * @apiDescription Send Password reset Code (OTP) to identifier of type loginType.
 * @apiName SendPasswordResetOTP
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (JSON Request Body) {String=usernames,emails,phones,facebook} loginType type of identifier to send OTP to.
 * @apiParam (JSON Request Body) {String} identifier The loginType's unique identifier for which to send password reset code to.
 *
 * @apiUse OTPStatus
 *
 */
func (s *handler) handleSendPassResetCode(w http.ResponseWriter, r *http.Request) {
	req := struct {
		LT     string `json:"loginType"`
		ToAddr string `json:"identifier"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, &req) {
		return
	}
	dbtStatus, err := s.auth.SendPassResetCode(req.LT, req.ToAddr)
	s.respondOn(w, r, req, NewDBTStatus(dbtStatus), http.StatusOK, err)
}

/**
 * @api {POST} /reset_password Reset password
 * @apiDescription Reset a user's password.
 * @apiName ResetPassword
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (JSON Request Body) {String=usernames,emails,phones,facebook} loginType type of identifier for which password reset is sort.
 * @apiParam (JSON Request Body) {String} identifier the loginType's unique identifier for which password reset is sort.
 * @apiParam (JSON Request Body) {String} OTP The password reset code sent to user during <a href="#api-Auth-SendPasswordResetOTP">SendPasswordResetOTP</a>.
 * @apiParam (JSON Request Body) {String} newSecret The new password.
 *
 * @apiUse VerifLogin
 *
 */
func (s *handler) handleResetPass(w http.ResponseWriter, r *http.Request) {
	req := struct {
		LT        string `json:"loginType"`
		OnAddress string `json:"identifier"`
		DBT       string `json:"OTP"`
		NewSecret string `json:"newSecret"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, &req) {
		return
	}
	vl, err := s.auth.SetPassword(req.LT, req.OnAddress, []byte(req.DBT), []byte(req.NewSecret))
	s.respondOn(w, r, req, NewVerifLogin(vl), http.StatusOK, err)
}

func (s *handler) handleError(w http.ResponseWriter, r *http.Request, reqData interface{}, err error) {
	reqDataB, _ := json.Marshal(reqData)
	log := r.Context().Value(ctxKeyLog).(logging.Logger).
		WithField(logging.FieldRequest, string(reqDataB))
	if s.auth.IsAuthError(err) || s.IsAuthError(err) {
		if s.auth.IsForbiddenError(err) || s.IsForbiddenError(err) {
			log.Warnf("Forbidden: %v", err)
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			log.Warnf("Unauthorized: %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}
		return
	}
	if s.auth.IsClientError(err) || s.IsClientError(err) {
		log.Warnf("Bad request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.auth.IsNotFoundError(err) || s.IsNotFoundError(err) {
		log.Warnf("Not found: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if s.auth.IsNotImplementedError(err) || s.IsNotImplementedError(err) {
		log.Warnf("Not implemented entity: %v", err)
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}
	log.Errorf("Internal error: %v", err)
	http.Error(w, internalErrorMessage, http.StatusInternalServerError)
}

func (s *handler) respondOn(w http.ResponseWriter, r *http.Request, reqData interface{}, respData interface{}, code int, err error) int {

	if err != nil {
		s.handleError(w, r, reqData, err)
		return 0
	}

	respBytes, err := json.Marshal(respData)
	if err != nil {
		s.handleError(w, r, reqData, err)
		return 0
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	i, err := w.Write(respBytes)
	if err != nil {
		log := r.Context().Value(ctxKeyLog).(logging.Logger)
		log.Errorf("unable write data to response stream: %v", err)
		return i
	}

	return i
}

func (s handler) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Nothing to see here", http.StatusNotFound)
}
