package rpc

import (
	"errors"
	"net/http"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/server/helper"
	"golang.org/x/net/context"
)

const (
	internalErrorMessage = "whoops! Something wicked happened"
)

type Logger interface {
	Error(interface{}, ...interface{}) error
	Warn(interface{}, ...interface{}) error
	Info(interface{}, ...interface{})
	Debug(interface{}, ...interface{})
	Fine(interface{}, ...interface{})
}

type Server struct {
	auth  *auth.Auth
	lg    Logger
	tIDCh chan int
	name  string
}

var ErrorNilAuth = errors.New("Auth cannot be nil")
var ErrorNilLogger = errors.New("Logger cannot be nil")
var ErrorEmptyName = errors.New("Name cannot be empty")

func New(name string, auth *auth.Auth, lg Logger) (*Server, error) {
	if auth == nil {
		return nil, ErrorNilAuth
	}
	if lg == nil {
		return nil, ErrorNilLogger
	}
	if name == "" {
		return nil, ErrorEmptyName
	}
	tIDCh := make(chan int)
	go helper.TransactionSerializer(tIDCh)
	return &Server{name: name, auth: auth, lg: lg, tIDCh: tIDCh}, nil
}

func (s *Server) Register(c context.Context, req *authms.RegisterRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - register user", tID)
	err := s.auth.Register(req.User, req.DeviceID, "")
	return s.respondOn(req.User, resp, http.StatusCreated, tID, err)
}

func (s *Server) LoginUserName(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - login user by username...", tID)
	authUsr, err := s.auth.LoginUserName(req.BasicID, req.Password,
		req.DeviceID, "")
	return s.respondOn(authUsr, resp, http.StatusOK, tID, err)
}

func (s *Server) LoginEmail(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	return errors.New("Not implemented")
}

func (s *Server) LoginPhone(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	return errors.New("Not implemented")
}

func (s *Server) LoginOAuth(c context.Context, req *authms.OAuthRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - login user by OAuth...", tID)
	authUsr, err := s.auth.LoginOAuth(req.OAuth, req.DeviceID, "")
	return s.respondOn(authUsr, resp, http.StatusOK, tID, err)
}

func (s *Server) UpdatePhone(c context.Context, req *authms.UpdateRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - update phone...", tID)
	err := s.auth.UpdatePhone(req.User, req.Token, req.DeviceID, "")
	return s.respondOn(req.User, resp, http.StatusOK, tID, err)
}

func (s *Server) UpdateOauth(c context.Context, req *authms.UpdateRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - update OAuth...", tID)
	err := s.auth.UpdateOAuth(req.User, req.AppName, req.Token, req.DeviceID, "")
	return s.respondOn(req.User, resp, http.StatusOK, tID, err)
}

func (s *Server) VerifyPhone(c context.Context, req *authms.SMSVerificationRequest, resp *authms.SMSVerificationResponse) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - verify phone...", tID)
	r, err := s.auth.VerifyPhone(req, "")
	return s.respondOnSMS(r, resp, tID, err)
}

func (s *Server) VerifyPhoneCode(c context.Context, req *authms.SMSVerificationCodeRequest, resp *authms.SMSVerificationResponse) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - verify phone...", tID)
	r, err := s.auth.VerifyPhoneCode(req, "")
	return s.respondOnSMS(r, resp, tID, err)
}

func (s *Server) respondOnSMS(r *authms.SMSVerificationStatus, resp *authms.SMSVerificationResponse, tID int, err error) error {
	if err != nil {
		if !auth.AuthError(err) {
			s.lg.Error("%d - internal auth error: %s", tID, err)
			resp.Detail = internalErrorMessage
			resp.Code = http.StatusInternalServerError
			return nil
		}
		// FIXME the error message may have TMI, sanitize
		resp.Detail = err.Error()
		resp.Code = http.StatusUnauthorized
		return nil
	}
	resp.Id = s.name
	resp.Code = http.StatusOK
	resp.Status = r
	return nil
}

func (s *Server) respondOn(authUsr *authms.User, resp *authms.Response, code int32, tID int, err error) error {
	if err != nil {
		if !auth.AuthError(err) {
			s.lg.Error("%d - internal auth error: %s", tID, err)
			resp.Detail = internalErrorMessage
			resp.Code = http.StatusInternalServerError
			return nil
		}
		// FIXME the error message may have TMI, sanitize
		resp.Detail = err.Error()
		resp.Code = http.StatusUnauthorized
		return nil
	}
	resp.Id = s.name
	resp.Code = code
	resp.User = authUsr
	return nil
}
