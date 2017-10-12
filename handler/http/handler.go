package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

type contextKey string

type Auth interface {
	IsNotImplementedError(error) bool
	IsClientError(error) bool
	IsForbiddenError(error) bool
	IsAuthError(error) bool
	RegisterSelf(loginType, userType, id string, secret []byte) (*model.User, error)
	RegisterSelfByLockedPhone(userType, devID, number string, password []byte) (*model.User, error)
	RegisterOther(JWT, newLoginType, userType, id, groupID string) (*model.User, error)
	UpdateIdentifier(JWT, loginType, newId string) (*model.User, error)
	UpdatePassword(JWT string, old, newPass []byte) error
	SetPassword(loginType, userID string, dbt, pass []byte) (*model.VerifLogin, error)
	SendVerCode(JWT, loginType, toAddr string) (*model.DBTStatus, error)
	SendPassResetCode(loginType, toAddr string) (*model.DBTStatus, error)
	VerifyAndExtendDBT(lt, usrID string, dbt []byte) (string, error)
	VerifyDBT(loginType, userID string, dbt []byte) (*model.VerifLogin, error)
	Login(loginType, identifier string, password []byte) (*model.User, error)
}

type Guard interface {
	APIKeyValid(key string) (string, error)
}

type handler struct {
	errors.NotImplErrCheck
	errors.AuthErrCheck
	errors.ClErrCheck
	auth   Auth
	guard  Guard
	logger logging.Logger
}

const (
	internalErrorMessage = "whoops! Something wicked happened"

	keyLoginType = "loginType"
	keySelfReg   = "selfReg"
	keyAPIKey    = "x-api-key"
	keyToken     = "token"
	keyDBT       = "dbt"
	keyExtend    = "extend"
	keyAddress   = "address"
	keyUserID    = "userID"

	ctxtKeyBody = contextKey("id")
	ctxKeyLog   = contextKey("log")

	valTrue   = "true"
	valDevice = "device"
)

func NewHandler(a Auth, g Guard, l logging.Logger) (http.Handler, error) {
	if a == nil {
		return nil, errors.New("Auth was nil")
	}
	if g == nil {
		return nil, errors.New("Guard was nil")
	}
	if l == nil {
		return nil, errors.New("Logger was nil")
	}

	r := mux.NewRouter().PathPrefix(config.WebRootURL).Subrouter()
	handler{auth: a, guard: g, logger: l}.handleRoute(r)

	return r, nil
}

func (s handler) handleRoute(r *mux.Router) {
	r.PathPrefix("/users/{" + keyUserID + "}/verify/{" + keyDBT + "}").
		Methods(http.MethodGet).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleVerifyCode)))

	r.PathPrefix("/{" + keyLoginType + "}/register").
		Methods(http.MethodPut).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleRegistration))))

	r.PathPrefix("/{" + keyLoginType + "}/verify/{" + keyAddress + "}").
		Methods(http.MethodGet).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleSendVerifCode)))

	r.PathPrefix("/{" + keyLoginType + "}/login").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleLogin))))

	r.PathPrefix("/{" + keyLoginType + "}/update").
		Methods(http.MethodPost).
		HandlerFunc(s.prepLogger(s.guardRoute(s.readReqBody(s.handleUpdate))))
	r.NotFoundHandler = http.HandlerFunc(s.prepLogger(s.notFoundHandler))
}

func (s handler) prepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := s.logger.WithField(logging.FieldTransID, uuid.New())
		log.WithFields(map[string]interface{}{
			logging.FieldURL:            r.URL,
			logging.FieldHost:           r.Host,
			logging.FieldMethod:         r.Method,
			logging.FieldRequestHandler: "HTTP",
			logging.FieldHttpReqObj:     r,
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

func (s *handler) handleRegistration(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &struct {
		UserType   string `json:"userType"`
		Identifier string `json:"identifier"`
		Secret     string `json:"secret"`
		GroupID    string `json:"groupID"`
		LT         string `json:"loginType"`
	}{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	vars := mux.Vars(r)
	req.LT = vars[keyLoginType]
	var usr *model.User
	var err error
	switch strings.ToLower(r.URL.Query().Get(keySelfReg)) {
	case valTrue:
		usr, err = s.auth.RegisterSelf(req.LT, req.UserType, req.Identifier, []byte(req.Secret))
	case valDevice:
		usr, err = s.auth.RegisterSelfByLockedPhone(req.LT, req.UserType, req.Identifier, []byte(req.Secret))
	default:
		JWT := r.URL.Query().Get(keyToken)
		usr, err = s.auth.RegisterOther(req.LT, JWT, req.UserType, req.Identifier, req.GroupID)
	}
	req.Secret = "" // prevent logging passwords.
	s.respondOn(w, r, req, usr, http.StatusCreated, err)
}

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
	authUsr, err := s.auth.Login(req.LT, req.Identifier, []byte(secret))
	s.respondOn(w, r, req, authUsr, http.StatusOK, err)
}

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
	s.respondOn(w, r, req, usr, http.StatusOK, err)
}

func (s *handler) handleSendVerifCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := struct {
		LT     string `json:"loginType"`
		ToAddr string `json:"address"`
		JWT    string `json:"token"`
	}{
		LT:     vars[keyLoginType],
		ToAddr: vars[keyAddress],
		JWT:    r.URL.Query().Get(keyToken),
	}
	resp, err := s.auth.SendVerCode(req.JWT, req.LT, req.ToAddr)
	s.respondOn(w, r, req, resp, http.StatusOK, err)
}

func (s *handler) handleVerifyCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	q := r.URL.Query()
	req := struct {
		UserID string `json:"userID"`
		LT     string `json:"loginType"`
		DBT    string `json:"dbt"`
		Extend string `json:"extend"`
	}{
		UserID: vars[keyUserID],
		LT:     q.Get(keyLoginType),
		DBT:    vars[keyDBT],
		Extend: q.Get(keyExtend),
	}
	var resp interface{}
	var err error
	if strings.EqualFold(req.Extend, valTrue) {
		resp, err = s.auth.VerifyAndExtendDBT(req.LT, req.UserID, []byte(req.DBT))
	} else {
		resp, err = s.auth.VerifyDBT(req.LT, req.UserID, []byte(req.DBT))
	}
	s.respondOn(w, r, req, resp, http.StatusOK, err)
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	response := struct {
		Status  int         `json:"status"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{
		Status: code,
		Data:   respData,
	}
	w.Header().Set("Content-Type", "application/json")
	respBytes, err := json.Marshal(&response)
	if err != nil {
		s.handleError(w, r, reqData, err)
		return 0
	}
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
	fmt.Fprint(w, "Nothing to see here")
}
