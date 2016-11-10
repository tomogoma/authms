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

//type Request authms.Request
//type History authms.History
type AppID authms.AppID

func (a AppID) Name() string { return a.AppName }

func (a AppID) UserID() string { return a.AppUserID }

func (a AppID) Validated() bool { return a.Verified }

//
//type Value authms.Value
//
//func (v Value) Value() string {
//	if v == nil {
//		return ""
//	}
//	return v.Value
//}
//
//func (v *Value) Validated() bool {
//	if v == nil {
//		return false
//	}
//	return v.Verified
//}

type User authms.User

func (u User) UserName() string { return u.Username }

//func (u User) Email() user.Valuer   { return u.Mail }
func (u User) EmailAddress() string { return u.Mail.Value }

//func (u User) Phone() user.Valuer   { return u.Phone }
func (u User) PhoneNumber() string { return u.Phone.Value }
func (u User) App() user.App       { return AppID(*u.AppID) }

type Server struct {
	auth  *auth.Auth
	lg    Logger
	tIDCh chan int
}

var ErrorNilAuth = errors.New("Auth cannot be nil")
var ErrorNilLogger = errors.New("Logger cannot be nil")

func New(auth *auth.Auth, lg Logger) (*Server, error) {
	if auth == nil {
		return nil, ErrorNilAuth
	}
	if lg == nil {
		return nil, ErrorNilLogger
	}
	tIDCh := make(chan int)
	go helper.TransactionSerializer(tIDCh)
	return &Server{auth: auth, lg: lg, tIDCh: tIDCh}, nil
}

func (s *Server) Register(c context.Context, req *authms.Request, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - register user...", tID)
	if req.User == nil {
		resp.Status = http.StatusBadRequest
		resp.Message = "Missing user Body"
		return nil
	}
	svdUsr, err := s.auth.RegisterUser(User(*req.User), req.User.Password, "", req.ForServiceID, req.RefererServiceID)
	if err != nil {
		s.lg.Fine("%d - check error is authentication or internal...", tID)
		if !auth.AuthError(err) {
			s.lg.Error("%d - registration error: %s", tID, err)
			resp.Status = http.StatusInternalServerError
			resp.Message = internalErrorMessage
			return nil
		}
		s.lg.Warn("%d - registration error: %s", tID, err)
		resp.Message = err.Error()
		resp.Status = http.StatusUnauthorized
		return nil
	}
	s.lg.Fine("%d - package registered user result...", tID)
	packageResponseUser(svdUsr, resp.User)
	resp.Status = http.StatusCreated
	s.lg.Info("%d - Registration complete", tID)
	return nil
}

func (s *Server) Login(c context.Context, req *authms.Request, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - login user...", tID)
	if req.User == nil {
		resp.Status = http.StatusBadRequest
		resp.Message = "Missing user Body"
		return nil
	}
	var authUsr user.User
	var err error
	if req.User.Username != "" {
		s.lg.Fine("%d - use username...", tID)
		authUsr, err = s.auth.LoginUserName(req.User.Username,
			req.User.Password, req.DeviceID, "", req.ForServiceID,
			req.RefererServiceID)
	} else if req.User.AppID != nil {
		s.lg.Fine("%d - use appID...", tID)
		authUsr, err = s.auth.LoginUserAppID(AppID(*req.User.AppID),
			req.User.Password, req.DeviceID, "", req.ForServiceID,
			req.RefererServiceID)
	}
	if err != nil {
		s.lg.Fine("%d - check error is authentication or internal...", tID)
		if !auth.AuthError(err) {
			s.lg.Error("%d - login error: %s", tID, err)
			resp.Message = internalErrorMessage
			resp.Status = http.StatusInternalServerError
			return nil
		}
		s.lg.Warn("%d - login error: %s", tID, err)
		resp.Message = "invalid userName/password combo or missing devID"
		resp.Status = http.StatusUnauthorized
		return nil
	}
	s.lg.Fine("%d - package authenticated user result...", tID)
	packageResponseUser(authUsr, resp.User)
	resp.Status = http.StatusOK
	s.lg.Info("%d - login complete", tID)
	return nil
}

func (s *Server) ValidateToken(c context.Context, req *authms.Request, resp *authms.Response) error {
	tID := <-s.tIDCh
	s.lg.Fine("%d - validate token...", tID)
	if req.User == nil {
		resp.Status = http.StatusBadRequest
		resp.Message = "Missing user Body"
		return nil
	}
	authUsr, err := s.auth.AuthenticateToken(req.User.Token, "",
		req.ForServiceID, req.RefererServiceID)
	if err != nil {
		s.lg.Fine("%d - check error is authentication or internal...", tID)
		if !auth.AuthError(err) {
			s.lg.Error("%d - token authentication error: %s", tID, err)
			resp.Message = internalErrorMessage
			resp.Status = http.StatusInternalServerError
			return nil
		}
		s.lg.Warn("%d - token authentication error: %s", tID, err)
		resp.Message = err.Error()
		resp.Status = http.StatusUnauthorized
		return nil
	}
	s.lg.Fine("%d - package authenticated user result...", tID)
	packageResponseUser(authUsr, resp.User)
	resp.Status = http.StatusOK
	s.lg.Info("%d - token validation complete", tID)
	return errors.New("Not yet implemented")
}

func packageResponseUser(rcv user.User, resp *authms.User) {
	resp.Id = int64(rcv.ID())
	resp.Username = rcv.UserName()
	resp.Token = rcv.Token("")
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
	resp.LoginHistory = prevLogins
	if e := rcv.Email(); e != nil && fmt.Sprintf("%v", e) != "<nil>" {
		resp.Mail = &authms.Value{
			Value:    e.Value(),
			Verified: e.Validated(),
		}
	}
	if p := rcv.Phone(); p != nil && fmt.Sprintf("%v", p) != "<nil>" {
		resp.Phone = &authms.Value{
			Value:    p.Value(),
			Verified: p.Validated(),
		}
	}
	if app := rcv.App(); app != nil && fmt.Sprintf("%v", app) != "<nil>" {
		resp.AppID = &authms.AppID{
			AppName:   app.Name(),
			AppUserID: app.UserID(),
			Verified:  app.Validated(),
		}
	}
}
