package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-typed-errors"
)

type contextKey string

type Auth interface {
	IsNotImplementedError(error) bool
	IsClientError(error) bool
	IsForbiddenError(error) bool
	IsAuthError(error) bool
	Register(user *authms.User, devID, rIP string) error
	LoginUserName(uName, pass, devID, rIP string) (*authms.User, error)
	LoginEmail(email, pass, devID, rIP string) (*authms.User, error)
	LoginPhone(phone, pass, devID, rIP string) (*authms.User, error)
	LoginOAuth(app *authms.OAuth, devID, rIP string) (*authms.User, error)
	UpdatePhone(user *authms.User, token, devID, rIP string) error
	UpdateOAuth(user *authms.User, appName, token, devID, rIP string) error
	VerifyPhone(req *authms.SMSVerificationRequest, rIP string) (*authms.SMSVerificationStatus, error)
	VerifyPhoneCode(req *authms.SMSVerificationCodeRequest, rIP string) (*authms.SMSVerificationStatus, error)
}

type Handler struct {
	errors.NotImplErrCheck
	errors.AuthErrCheck
	errors.ClErrCheck
	auth Auth
}

const (
	internalErrorMessage = "whoops! Something wicked happened"

	urlVarLoginType = "loginType"

	authTypePhone    = "phones"
	authTypeOAuth    = "oauths"
	authTypeEmail    = "emails"
	authTypeUsername = "usernames"
	authTypeCode     = "codes"

	ctxtKeyBody = contextKey("id")
	ctxKeyLog   = contextKey("log")
)

func NewHandler(a Auth) (*Handler, error) {
	if a == nil {
		return nil, errors.New("auth was nil")
	}
	return &Handler{auth: a}, nil
}

func (s *Handler) HandleRoute(r *mux.Router) error {
	if r == nil {
		return errors.New("Router was nil")
	}
	r.PathPrefix("/register").
		Methods(http.MethodPost).
		HandlerFunc(s.readReqBody(s.handleRegistration))
	r.PathPrefix("/" + authTypeOAuth + "/login").
		Methods(http.MethodPost).
		HandlerFunc(s.readReqBody(s.handleLoginOAuth))
	r.PathPrefix("/" + authTypePhone + "/verify").
		Methods(http.MethodPost).
		HandlerFunc(s.readReqBody(s.handleVerifyPhone))
	r.PathPrefix("/" + authTypeCode + "/verify").
		Methods(http.MethodPost).
		HandlerFunc(s.readReqBody(s.handleVerifyCode))
	r.PathPrefix("/{" + urlVarLoginType + "}/login").
		Methods(http.MethodPost).
		HandlerFunc(s.readReqBody(s.handleLogin))
	r.PathPrefix("/{" + urlVarLoginType + "}/update").
		Methods(http.MethodPost).
		HandlerFunc(s.readReqBody(s.handleUpdate))
	return nil
}

func prepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logrus.WithField(logging.FieldTransID, uuid.New())
		log.WithFields(logrus.Fields{
			logging.FieldURL:            r.URL,
			logging.FieldHost:           r.Host,
			logging.FieldMethod:         r.Method,
			logging.FieldRequestHandler: "HTTP",
		}).Info("new request")
		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Handler) readReqBody(next http.HandlerFunc) http.HandlerFunc {
	return prepLogger(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		dataB, err := ioutil.ReadAll(r.Body)
		if err != nil {
			err = errors.NewClientf("Failed to read request body: %v", err)
			s.handleError(w, r, nil, err)
			return
		}
		ctx := context.WithValue(r.Context(), ctxtKeyBody, dataB)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// unmarshalJSONOrRespondError returns true if json is extracted from
// data into req successfully, otherwise, it writes an error response into
// w and returns false.
// The Context in r should contain a logrus Entry with key ctxKeyLog
// for logging in case of error
func (s *Handler) unmarshalJSONOrRespondError(w http.ResponseWriter, r *http.Request, data []byte, req interface{}) bool {
	err := json.Unmarshal(data, req)
	if err != nil {
		err = errors.NewClientf("failed to unmarshal JSON request from body: %v", err)
		s.handleError(w, r, nil, err)
		return false
	}
	return true
}

func (s *Handler) handleRegistration(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.RegisterRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	err := s.auth.Register(req.User, req.DeviceID, r.RemoteAddr)
	// todo req may have been mutated and thus not good enough for debugging!
	s.respondOn(w, r, req, req.User, http.StatusCreated, err)
}

func (s *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.BasicAuthRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	vars := mux.Vars(r)
	var authUsr *authms.User
	var err error
	switch vars[urlVarLoginType] {
	case authTypeUsername:
		authUsr, err = s.auth.LoginUserName(req.BasicID, req.Password,
			req.DeviceID, "")
	case authTypeEmail:
		authUsr, err = s.auth.LoginEmail(req.BasicID, req.Password,
			req.DeviceID, "")
	case authTypePhone:
		authUsr, err = s.auth.LoginPhone(req.BasicID, req.Password,
			req.DeviceID, "")
	default:
		http.Error(w, "unknown login type", http.StatusBadRequest)
		return
	}
	s.respondOn(w, r, req, authUsr, http.StatusOK, err)
}

func (s *Handler) handleLoginOAuth(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.OAuthRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	authUsr, err := s.auth.LoginOAuth(req.OAuth, req.DeviceID, "")
	s.respondOn(w, r, req, authUsr, http.StatusOK, err)
}

func (s *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.UpdateRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	vars := mux.Vars(r)
	var err error
	switch vars[urlVarLoginType] {
	case authTypePhone:
		err = s.auth.UpdatePhone(req.User, req.Token, req.DeviceID, "")
	case authTypeOAuth:
		err = s.auth.UpdateOAuth(req.User, req.AppName, req.Token, req.DeviceID, "")
	case authTypeUsername:
		fallthrough
	case authTypeEmail:
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	default:
		s.handleError(w, r, req, errors.NewClient("Unknown auth type"))
		return
	}
	// todo req may have been mutated and thus not good enough for debugging!
	s.respondOn(w, r, req, req.User, http.StatusOK, err)
}

func (s *Handler) handleVerifyPhone(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.SMSVerificationRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	resp, err := s.auth.VerifyPhone(req, "")
	s.respondOn(w, r, req, resp, http.StatusOK, err)
}

func (s *Handler) handleVerifyCode(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.SMSVerificationCodeRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	resp, err := s.auth.VerifyPhoneCode(req, "")
	s.respondOn(w, r, req, resp, http.StatusOK, err)
}

func (s *Handler) handleError(w http.ResponseWriter, r *http.Request, reqData interface{}, err error) {
	log := r.Context().Value(ctxKeyLog).(*logrus.Entry).WithField(logging.FieldRequest, reqData)
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

func (s *Handler) respondOn(w http.ResponseWriter, r *http.Request, reqData interface{}, respData interface{}, code int, err error) int {
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
		log := r.Context().Value(ctxKeyLog).(*logrus.Entry)
		log.Errorf("unable write data to response stream: %v", err)
		return i
	}
	return i
}
