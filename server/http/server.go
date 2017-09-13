package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/errors"
)

type contextKey string

const (
	internalErrorMessage = "whoops! Something wicked happened"

	urlVarLoginType = "loginType"

	authTypePhone    = "phones"
	authTypeOAuth    = "oauths"
	authTypeEmail    = "emails"
	authTypeUsername = "usernames"
	authTypeCode     = "codes"

	logKeyTransID = "transactionID"
	logKeyURL     = "url"
	logKeyHost    = "host"
	logKeyMethod  = "method"
	logKeyRequest = "request"

	ctxtKeyBody = contextKey("id")
	ctxKeyLog   = contextKey("log")
)

type Server struct {
	errors.AuthErrCheck
	errors.ClErrCheck
	auth *auth.Auth
}

func New(auth *auth.Auth) (*Server, error) {
	if auth == nil {
		return nil, errors.New("auth was nil")
	}
	return &Server{auth: auth}, nil
}

func (s *Server) HandleRoute(r *mux.Router) {
	r.PathPrefix("/register").
		Methods(http.MethodPost).
		HandlerFunc(s.readReqBody(s.handleRegistration))
	r.PathPrefix("/" + authTypeOAuth + "/login").
		Methods(http.MethodPost).
		HandlerFunc(s.readReqBody(s.handleLoginOAuth))
	r.PathPrefix("/" + authTypePhone + "/verify/").
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
}

func prepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logrus.WithFields(logrus.Fields{
			logKeyTransID: uuid.New(),
			logKeyURL:     r.URL,
			logKeyHost:    r.Host,
			logKeyMethod:  r.Method,
		})
		log.Info("new request")
		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *Server) readReqBody(next http.HandlerFunc) http.HandlerFunc {
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
func (s *Server) unmarshalJSONOrRespondError(w http.ResponseWriter, r *http.Request, data []byte, req interface{}) bool {
	err := json.Unmarshal(data, req)
	if err != nil {
		err = errors.NewClientf("failed to unmarshal JSON request from body: %v", err)
		s.handleError(w, r, nil, err)
		return false
	}
	return true
}

func (s *Server) handleRegistration(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.RegisterRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	err := s.auth.Register(req.User, req.DeviceID, r.RemoteAddr)
	// todo req may have been mutated and thus not good enough for debugging!
	s.respondOn(w, r, req, req.User, http.StatusCreated, err)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleLoginOAuth(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.OAuthRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	authUsr, err := s.auth.LoginOAuth(req.OAuth, req.DeviceID, "")
	s.respondOn(w, r, req, authUsr, http.StatusOK, err)
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleVerifyPhone(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.SMSVerificationRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	resp, err := s.auth.VerifyPhone(req, "")
	s.respondOn(w, r, req, resp, http.StatusOK, err)
}

func (s *Server) handleVerifyCode(w http.ResponseWriter, r *http.Request) {
	dataB := r.Context().Value(ctxtKeyBody).([]byte)
	req := &authms.SMSVerificationCodeRequest{}
	if !s.unmarshalJSONOrRespondError(w, r, dataB, req) {
		return
	}
	resp, err := s.auth.VerifyPhoneCode(req, "")
	s.respondOn(w, r, req, resp, http.StatusOK, err)
}

func (s *Server) handleError(w http.ResponseWriter, r *http.Request, reqData interface{}, err error) {
	log := r.Context().Value(ctxKeyLog).(*logrus.Entry).WithField(logKeyRequest, reqData)
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
	log.Errorf("Internal error: %v", err)
	http.Error(w, internalErrorMessage, http.StatusInternalServerError)
}

func (s *Server) respondOn(w http.ResponseWriter, r *http.Request, reqData interface{}, respData interface{}, code int, err error) int {
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
