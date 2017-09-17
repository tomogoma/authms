package rpc

import (
	"errors"
	"net/http"

	"github.com/micro/go-micro/server"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/proto/authms"
	"golang.org/x/net/context"
)

type Server struct {
	auth *model.Auth
	name string
}

type ctxKey string

const (
	internalErrorMessage = "whoops! Something wicked happened"

	ctxKeyLog = "log"
)

func New(name string, auth *model.Auth) (*Server, error) {
	if auth == nil {
		return nil, errors.New("nil auth")
	}
	if name == "" {
		return nil, errors.New("empty name")
	}
	return &Server{name: name, auth: auth}, nil
}

func LogWrapper(next server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		log := logrus.WithField(logging.FieldTransID, uuid.New())
		log.WithFields(logrus.Fields{
			logging.FieldTransID:        uuid.New(),
			logging.FieldService:        req.Service(),
			logging.FieldMethod:         req.Method(),
			logging.FieldRequestHandler: "RPC",
		}).Info("new request")
		ctx = context.WithValue(ctx, ctxKeyLog, log)
		return next(ctx, req, rsp)
	}
}

func (s *Server) Wrapper(next server.HandlerFunc) server.HandlerFunc {
	return LogWrapper(next)
}

func (s *Server) Register(c context.Context, req *authms.RegisterRequest, resp *authms.Response) error {
	err := s.auth.Register(req.User, req.DeviceID, "")
	return s.respondOnUser(c, req.User, resp, http.StatusCreated, err)
}

func (s *Server) LoginUserName(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	authUsr, err := s.auth.LoginUserName(req.BasicID, req.Password,
		req.DeviceID, "")
	return s.respondOnUser(c, authUsr, resp, http.StatusOK, err)
}

func (s *Server) LoginEmail(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	authUsr, err := s.auth.LoginEmail(req.BasicID, req.Password,
		req.DeviceID, "")
	return s.respondOnUser(c, authUsr, resp, http.StatusOK, err)
}

func (s *Server) LoginPhone(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	authUsr, err := s.auth.LoginPhone(req.BasicID, req.Password,
		req.DeviceID, "")
	return s.respondOnUser(c, authUsr, resp, http.StatusOK, err)
}

func (s *Server) LoginOAuth(c context.Context, req *authms.OAuthRequest, resp *authms.Response) error {
	authUsr, err := s.auth.LoginOAuth(req.OAuth, req.DeviceID, "")
	return s.respondOnUser(c, authUsr, resp, http.StatusOK, err)
}

func (s *Server) UpdatePhone(c context.Context, req *authms.UpdateRequest, resp *authms.Response) error {
	err := s.auth.UpdatePhone(req.User, req.Token, req.DeviceID, "")
	return s.respondOnUser(c, req.User, resp, http.StatusOK, err)
}

func (s *Server) UpdateOauth(c context.Context, req *authms.UpdateRequest, resp *authms.Response) error {
	err := s.auth.UpdateOAuth(req.User, req.AppName, req.Token, req.DeviceID, "")
	return s.respondOnUser(c, req.User, resp, http.StatusOK, err)
}

func (s *Server) VerifyPhone(c context.Context, req *authms.SMSVerificationRequest, resp *authms.SMSVerificationResponse) error {
	r, err := s.auth.VerifyPhone(req, "")
	return s.respondOnSMS(c, r, resp, err)
}

func (s *Server) VerifyPhoneCode(c context.Context, req *authms.SMSVerificationCodeRequest, resp *authms.SMSVerificationResponse) error {
	r, err := s.auth.VerifyPhoneCode(req, "")
	return s.respondOnSMS(c, r, resp, err)
}

func (s *Server) respondOnSMS(ctx context.Context, r *authms.SMSVerificationStatus, resp *authms.SMSVerificationResponse, err error) error {
	if err != nil {
		log := ctx.Value(ctxKeyLog).(*logrus.Entry)
		if s.auth.IsAuthError(err) {
			log.Warnf("Unauthorized: %v", err)
			resp.Detail = err.Error()
			resp.Code = http.StatusUnauthorized
			return nil
		}
		if s.auth.IsClientError(err) {
			log.Warnf("Bad request: %v", err)
			resp.Detail = err.Error()
			resp.Code = http.StatusBadRequest
			return nil
		}
		if s.auth.IsNotImplementedError(err) {
			log.Warnf("Not implemented: %v", err)
			resp.Detail = "This feature is not available"
			resp.Code = http.StatusNotImplemented
			return nil
		}
		log.Errorf("Internal error: %v", err)
		resp.Detail = internalErrorMessage
		resp.Code = http.StatusInternalServerError
		return nil
	}
	resp.Id = s.name
	resp.Code = http.StatusOK
	resp.Status = r
	return nil
}

func (s *Server) respondOnUser(ctx context.Context, authUsr *authms.User, resp *authms.Response, code int32, err error) error {
	if err != nil {
		log := ctx.Value(ctxKeyLog).(*logrus.Entry)
		if s.auth.IsAuthError(err) {
			log.Warnf("Unauthorized: %v", err)
			resp.Detail = err.Error()
			resp.Code = http.StatusUnauthorized
			return nil
		}
		if s.auth.IsClientError(err) {
			log.Warnf("Bad request: %v", err)
			resp.Detail = err.Error()
			resp.Code = http.StatusBadRequest
			return nil
		}
		if s.auth.IsNotImplementedError(err) {
			log.Warnf("Not implemented: %v", err)
			resp.Detail = "This feature is not available"
			resp.Code = http.StatusNotImplemented
			return nil
		}
		log.Errorf("Internal error: %v", err)
		resp.Detail = internalErrorMessage
		resp.Code = http.StatusInternalServerError
		return nil
	}
	resp.Id = s.name
	resp.Code = code
	resp.User = authUsr
	return nil
}
