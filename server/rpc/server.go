package rpc

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/model/history"
	"github.com/tomogoma/authms/auth/model/user"
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

type OAuthRequest authms.OAuthRequest

func (a OAuthRequest) Name() string {
	return a.AppName
}
func (a OAuthRequest) UserID() string {
	return a.AppUserID
}
func (a OAuthRequest) Validated() bool {
	return false
}
func (a OAuthRequest) Token() string {
	return a.AppToken
}

type OAuth authms.OAuth

func (a OAuth) Name() string {
	return a.AppName
}
func (a OAuth) UserID() string {
	return a.AppUserID
}
func (a OAuth) Validated() bool {
	return false
}
func (a OAuth) Token() string {
	return a.AppToken
}

type User authms.User

func (u User) UserName() string {
	return u.Username
}
func (u User) EmailAddress() string {
	return u.Mail.Value
}
func (u User) PhoneNumber() string {
	return u.Phone.Value
}
func (u User) App() user.App {
	return OAuth(*u.OAuth)
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

func (s *Server) RegisterOAuth(c context.Context, req *authms.OAuthRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - register user by oauth id...", tID)
	svdUsr, err := s.auth.RegisterOAuth(OAuthRequest((*req)))
	return s.respondOn(svdUsr, resp, http.StatusCreated, tID, err)
}

func (s *Server) RegisterUserName(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - register user by username...", tID)
	svdUsr, err := s.auth.RegisterUserName(req.BasicID, req.Password)
	return s.respondOn(svdUsr, resp, http.StatusCreated, tID, err)
}

func (s *Server) RegisterEmail(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - register user by email...", tID)
	svdUsr, err := s.auth.RegisterEmail(req.BasicID, req.Password)
	return s.respondOn(svdUsr, resp, http.StatusCreated, tID, err)
}

func (s *Server) RegisterPhone(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - register user by phone...", tID)
	svdUsr, err := s.auth.RegisterPhone(req.BasicID, req.Password)
	return s.respondOn(svdUsr, resp, http.StatusCreated, tID, err)
}

func (s *Server) LoginUserName(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - login user by username...", tID)
	authUsr, err := s.auth.LoginUserName(req.BasicID, req.Password,
		req.DeviceID, "", req.ForServiceID, req.RefererServiceID)
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
	authUsr, err := s.auth.LoginOAuth(OAuthRequest(*req), req.DeviceID, "",
		req.ForServiceID, req.RefererServiceID)
	return s.respondOn(authUsr, resp, http.StatusOK, tID, err)
}

func (s *Server) ValidateToken(c context.Context, req *authms.TokenRequest, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - validate token...", tID)
	authUsr, err := s.auth.AuthenticateToken(req.Token, "",
		req.ForServiceID, req.RefererServiceID)
	return s.respondOn(authUsr, resp, http.StatusOK, tID, err)
}

func (s *Server) respondOn(authUsr user.User, resp *authms.Response, code int32, tID int, err error) error {
	if err != nil {
		if !auth.AuthError(err) {
			s.lg.Error("%d - internal auth error: %s", tID, err)
			resp.Detail = internalErrorMessage
			resp.Code = http.StatusInternalServerError
			return nil
		}
		s.lg.Debug("%d - auth error: %s", tID, err)
		// FIXME the error message may have TMI, sanitize
		resp.Detail = err.Error()
		resp.Code = http.StatusUnauthorized
		return nil
	}
	s.lg.Fine("%d - package auth result...", tID)
	s.packageResponseUser(code, authUsr, resp)
	s.lg.Info("%d - auth complete", tID)
	return nil
}

func (s *Server) packageResponseUser(status int32, rcv user.User, resp *authms.Response) {
	resp.Id = s.name
	resp.Code = status
	if resp.User == nil {
		resp.User = new(authms.User)
	}
	resp.User.Id = int64(rcv.ID())
	resp.User.Username = rcv.UserName()
	resp.User.Token = rcv.Token("")
	prevLogins := make([]*authms.History, len(rcv.PreviousLogins()))
	for i, h := range rcv.PreviousLogins() {
		prevLogins[i] = &authms.History{
			Id:            int64(h.ID()),
			UserID:        int64(h.UserID()),
			IpAddress:     h.IPAddress(),
			Date:          h.Date().Format(time.RFC3339),
			AccessType:    history.DecodeAccessMethod(h.AccessMethod()),
			SuccessStatus: h.Successful(),
		}
	}
	resp.User.LoginHistory = prevLogins
	if e := rcv.Email(); e != nil && fmt.Sprintf("%v", e) != "<nil>" {
		resp.User.Mail = &authms.Value{
			Value:    e.Value(),
			Verified: e.Validated(),
		}
	}
	if p := rcv.Phone(); p != nil && fmt.Sprintf("%v", p) != "<nil>" {
		resp.User.Phone = &authms.Value{
			Value:    p.Value(),
			Verified: p.Validated(),
		}
	}
	if app := rcv.App(); app != nil && fmt.Sprintf("%v", app) != "<nil>" {
		resp.User.OAuth = &authms.OAuth{
			AppName:   app.Name(),
			AppUserID: app.UserID(),
		}
	}
}
