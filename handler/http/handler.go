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
	IsNotFoundError(error) bool
	IsNotImplementedError(error) bool
	IsClientError(error) bool
	IsForbiddenError(error) bool
	IsAuthError(error) bool
	RegisterFirst(loginType, userType, id string, secret []byte) (*model.User, error)
	CanRegisterFirst() (bool, error)
	RegisterSelf(loginType, userType, id string, secret []byte) (*model.User, error)
	RegisterSelfByLockedPhone(userType, devID, number string, password []byte) (*model.User, error)
	RegisterOther(JWT, newLoginType, userType, id, groupID string) (*model.User, error)
	UpdateIdentifier(JWT, loginType, newId string) (*model.User, error)
	UpdatePassword(JWT string, old, newPass []byte) error
	SetPassword(loginType, onAddr string, dbt, pass []byte) (*model.VerifLogin, error)
	SendVerCode(JWT, loginType, toAddr string) (*model.DBTStatus, error)
	SendPassResetCode(loginType, toAddr string) (*model.DBTStatus, error)
	VerifyAndExtendDBT(lt, forAddr string, dbt []byte) (string, error)
	VerifyDBT(loginType, forAddr string, dbt []byte) (*model.VerifLogin, error)
	Login(loginType, identifier string, password []byte) (*model.User, error)
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

	keyLoginType  = "loginType"
	keySelfReg    = "selfReg"
	keyAPIKey     = "x-api-key"
	keyToken      = "token"
	keyOTP        = "OTP"
	keyExtend     = "extend"
	keyIdentifier = "identifier"
	keyUserID     = "userID"

	ctxtKeyBody = contextKey("id")
	ctxKeyLog   = contextKey("log")

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
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleRegisterFirst))))

	r.PathPrefix("/reset_password/send_otp").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleSendPassResetCode))))

	r.PathPrefix("/reset_password").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleResetPass))))

	r.PathPrefix("/{" + keyLoginType + "}/register").
		Methods(http.MethodPut).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleRegistration))))

	r.PathPrefix("/{" + keyLoginType + "}/verify/{" + keyOTP + "}").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleVerifyCode))))

	r.PathPrefix("/{" + keyLoginType + "}/verify").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleSendVerifCode))))

	r.PathPrefix("/{" + keyLoginType + "}/login").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleLogin))))

	r.PathPrefix("/{" + keyLoginType + "}/update").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleUpdate))))

	r.PathPrefix("/" + config.DocsPath).
		Handler(http.FileServer(http.Dir(config.DefaultDocsDir())))

	r.NotFoundHandler = http.HandlerFunc(s.prepLogger(s.notFoundHandler))
}

func (s handler) prepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log := s.logger.WithHTTPRequest(r).
			WithField(logging.FieldTransID, uuid.New())

		log.WithFields(map[string]interface{}{
			logging.FieldURL:            r.URL,
			logging.FieldHost:           r.Host,
			logging.FieldMethod:         r.Method,
			logging.FieldRequestHandler: "HTTP",
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

func (s *handler) readReqBody(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		dataB, err := ioutil.ReadAll(r.Body)
		if err != nil {
			err = errors.NewClientf("Failed to read request body: %v", err)
			s.handleError(w, r, nil, err)
			return
		}
		ctx := context.WithValue(r.Context(), ctxtKeyBody, dataB)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// unmarshalJSONOrRespondError returns true if json is extracted from
// data into req successfully, otherwise, it writes an error response into
// w and returns false.
// The Context in r should contain a logging.Logger with key ctxKeyLog
// for logging in case of error
func (s *handler) unmarshalJSONOrRespondError(w http.ResponseWriter, r *http.Request, data []byte, req interface{}) bool {
	err := json.Unmarshal(data, req)
	if err != nil {
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
 * @apiSuccess (200) {String} name Micro-service name.
 * @apiSuccess (200)  {String} version Current running version.
 * @apiSuccess (200)  {String} description Short description of the micro-service.
 * @apiSuccess (200)  {String} canonicalName Canonical name of the micro-service.
 * @apiSuccess (200)  {String} needRegSuper true if a super-user has been registered, false otherwise.
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
 * @api {put} /first_user First User
 * @apiDescription Register the first super-user (super admin)
 * @apiName FirstUser
 * @apiVersion 0.1.0
 * @apiGroup Setup
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam {Enum} userType Type of user [individual|company]
 * @apiParam {Enum} loginType Type of identifier [usernames|emails|phones|facebook]
 * @apiParam {String} identifier The 'username' corresponding to loginType
 * @apiParam {String} secret The users password
 *
 * @apiSuccess (201) {Object} json-body See <a href="#api-Objects-User">User</a>.
 *
 */
func (s *handler) handleRegisterFirst(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &struct {
		UserType   string `json:"userType"`
		LT         string `json:"loginType"`
		Identifier string `json:"identifier"`
		Secret     string `json:"secret"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	usr, err := s.auth.RegisterFirst(req.LT, req.UserType, req.Identifier, []byte(req.Secret))
	req.Secret = "" // prevent logging passwords.
	s.respondOn(w, r, req, NewUser(usr), http.StatusCreated, err)
}

/**
 * @api {put} /:loginType/register?selfReg=:selfReg Register
 * @apiDescription  Register new user.
 * Registration can be:
 * - self registration - provide URL param selfReg=true
 * - self registration by unique device ID - provide URL param selfReg=device
 * - or other user (by admin) - don't provide URL params
 *
 * loginType is what the user will be logging in by, can be one of:
 * - usernames
 * - emails
 * - phones
 * - facebook
 *
 * @apiName Register
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam {Enum} userType Type of user [individual|company]
 * @apiParam {String} identifier The 'username' corresponding to loginType
 * @apiParam {String} secret The users password
 * @apiParam {String} groupID [only if selfReg not set or false] groupID to add this user to
 * @apiParam {String} deviceID [only if selfReg set to device] the unique device ID for the user
 *
 * @apiSuccess (201) {Object} json-body See <a href="#api-Objects-User">User</a>.
 *
 */
func (s *handler) handleRegistration(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &struct {
		UserType   string `json:"userType"`
		Identifier string `json:"identifier"`
		Secret     string `json:"secret"`
		GroupID    string `json:"groupID"`
		DevID      string `json:"deviceID"`
		LT         string `json:"loginType"`
		SelfReg    string `json:"selfReg"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
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
 * @apiHeader Authorization Basic auth containing identifier/secret, both provided during <a href="#api-Auth-Register">Registration</a>
 *
 * @apiSuccess (200) {Object} json-body See <a href="#api-Objects-User">User</a>.
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
 * @api {POST} /:loginType/update?token=:JWT Update Identifier
 * @apiDescription Update (or set for first time) the identifier details for loginType.
 * See <a href="#api-Auth-Register">Register</a> for loginType.
 * See <a href="#api-Objects-User">User</a> for how to access the JWT.
 * @apiName UpdateIdentifier
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam {String} identifier The new 'username' corresponding to loginType
 *
 * @apiSuccess (200) {Object} json-body See <a href="#api-Objects-User">User</a>.
 *
 */
func (s *handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &struct {
		Identifier string `json:"identifier"`
		LT         string `json:"loginType"`
		JWT        string `json:"token"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	vars := mux.Vars(r)
	req.LT = vars[keyLoginType]
	req.JWT = r.URL.Query().Get(keyToken)
	usr, err := s.auth.UpdateIdentifier(req.JWT, req.LT, req.Identifier)
	s.respondOn(w, r, req, NewUser(usr), http.StatusOK, err)
}

/**
 * @api {POST} /:loginType/verify?token=:JWT Send Verification Code
 * @apiDescription Send OTP to identifier of type loginType for purpose of verifying identifier.
 * See <a href="#api-Auth-Register">Register</a> for loginType options.
 * @apiName SendVerificationCode
 * @apiVersion 0.2.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam {String} identifier The loginType's address to be verified.
 *
 * @apiSuccess (200) {Object} json-body See <a href="#api-Objects-OTPStatus">OTPStatus</a>.
 *
 */
func (s *handler) handleSendVerifCode(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &struct {
		LT     string `json:"loginType"`
		ToAddr string `json:"identifier"`
		JWT    string `json:"token"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	vars := mux.Vars(r)
	req.LT = vars[keyLoginType]
	req.JWT = r.URL.Query().Get(keyToken)
	dbtStatus, err := s.auth.SendVerCode(req.JWT, req.LT, req.ToAddr)
	s.respondOn(w, r, req, NewDBTStatus(dbtStatus), http.StatusOK, err)
}

/**
 * @api {POST} /:loginType/verify/:OTP?extend=:extend Verify OTP
 * @apiDescription Verify OTP.
 * See <a href="#api-Auth-Register">Register</a> for loginType options.
 * extend can be set to "true" if intent on extending the expiry of the OTP.
 * @apiName VerifyOTP
 * @apiVersion 0.2.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam {String} identifier The loginType's address to whom the OTP was sent.
 *
 * @apiSuccess (200) {String} OTP [if extending OTP] the new OTP with extended expiry
 *
 * @apiSuccess (200) {Object} json-body [if not extending OTP] see <a href="#api-Objects-VerifLogin">VerifLogin</a>.
 *
 */
func (s *handler) handleVerifyCode(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &struct {
		Identifier string `json:"identifier"`
		LT         string `json:"loginType"`
		DBT        string `json:"OTP"`
		Extend     string `json:"extend"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
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
 * See <a href="#api-Auth-Register">Register</a> for loginType and identifier options.
 * @apiName SendPasswordResetOTP
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam {String} loginType See <a href="#api-Auth-Register">Register</a> for loginType options.
 * @apiParam {String} identifier See <a href="#api-Auth-Register">Register</a> for identifier options.
 *
 * @apiSuccess (200) {Object} json-body See <a href="#api-Objects-OTPStatus">OTPStatus</a>.
 *
 */
func (s *handler) handleSendPassResetCode(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := struct {
		LT     string `json:"loginType"`
		ToAddr string `json:"identifier"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, &req) {
		return
	}
	dbtStatus, err := s.auth.SendPassResetCode(req.LT, req.ToAddr)
	s.respondOn(w, r, req, NewDBTStatus(dbtStatus), http.StatusOK, err)
}

/**
 * @api {POST} /reset_password Reset password
 * @apiDescription Send Password reset Code (OTP) to identifier of type loginType.
 * See <a href="#api-Auth-Register">Register</a> for loginType and identifier options.
 * @apiName ResetPassword
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam {String} loginType See <a href="#api-Auth-Register">Register</a> for loginType options.
 * @apiParam {String} identifier See <a href="#api-Auth-Register">Register</a> for identifier options.
 * @apiParam {String} OTP The password reset code sent to user during <a href="#api-Auth-SendPasswordResetOTP">SendPasswordResetOTP</a>.
 * @apiParam {String} newSecret The new password.
 *
 * @apiSuccess (200) {Object} json-body See <a href="#api-Objects-VerifLogin">VerifLogin</a>.
 *
 */
func (s *handler) handleResetPass(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := struct {
		LT        string `json:"loginType"`
		OnAddress string `json:"identifier"`
		DBT       string `json:"OTP"`
		NewSecret string `json:"newSecret"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, &req) {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
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
