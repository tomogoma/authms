package rpc

import (
	"net/http"
	"strconv"
	"time"
	"github.com/micro/go-micro/server"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/proto/authms"
	"golang.org/x/net/context"
	"github.com/tomogoma/go-commons/errors"
)

type Handler struct {
	auth *model.Authentication
	name string
}

type ctxKey string

const (
	internalErrorMessage = "whoops! Something wicked happened"

	ctxKeyLog = "log"
)

func NewHandler(name string, auth *model.Authentication) (*Handler, error) {
	if auth == nil {
		return nil, errors.New("nil auth")
	}
	if name == "" {
		return nil, errors.New("empty name")
	}
	return &Handler{name: name, auth: auth}, nil
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

func (s *Handler) Wrapper(next server.HandlerFunc) server.HandlerFunc {
	return LogWrapper(next)
}

func (s *Handler) Register(c context.Context, req *authms.RegisterRequest, resp *authms.Response) error {
	if req.User == nil {
		resp.Code = http.StatusBadRequest
		resp.Detail = "Missing user"
	}
	var usr *model.User
	var err error
	if req.User.UserName != "" {
		_, err = s.auth.RegisterSelf(model.LoginTypeUsername, model.UserTypeIndividual, req.User.UserName, []byte(req.User.Password))
		if err != nil {
			return s.respondOnUser(c, usr, resp, http.StatusCreated, err)
		}
		usr, err = s.auth.Login(model.LoginTypeUsername, req.User.UserName, []byte(req.User.Password))
	} else if req.User.Phone != nil {
		_, err = s.auth.RegisterSelf(model.LoginTypePhone, model.UserTypeIndividual, req.User.Phone.Value, []byte(req.User.Password))
		if err != nil {
			return s.respondOnUser(c, usr, resp, http.StatusCreated, err)
		}
		usr, err = s.auth.Login(model.LoginTypePhone, req.User.Phone.Value, []byte(req.User.Password))
	} else if req.User.Email != nil {
		_, err = s.auth.RegisterSelf(model.LoginTypeEmail, model.UserTypeIndividual, req.User.Email.Value, []byte(req.User.Password))
		if err != nil {
			return s.respondOnUser(c, usr, resp, http.StatusCreated, err)
		}
		usr, err = s.auth.Login(model.LoginTypeEmail, req.User.Email.Value, []byte(req.User.Password))
	} else if req.User.OAuths != nil {
		if fb := req.User.OAuths["facebook"]; fb != nil {
			usr, err = s.auth.RegisterSelf(model.LoginTypeFacebook, model.UserTypeIndividual, fb.AppToken, nil)
			if err != nil {
				return s.respondOnUser(c, usr, resp, http.StatusCreated, err)
			}
			usr, err = s.auth.Login(model.LoginTypeFacebook, fb.AppToken, nil)
		}
	} else {
		err := errors.NewClient("no login type found")
		return s.respondOnUser(c, nil, resp, http.StatusCreated, err)
	}
	return s.respondOnUser(c, usr, resp, http.StatusCreated, err)
}

func (s *Handler) LoginUserName(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	authUsr, err := s.auth.Login(model.LoginTypeUsername, req.BasicID, []byte(req.Password))
	return s.respondOnUser(c, authUsr, resp, http.StatusOK, err)
}

func (s *Handler) LoginEmail(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	authUsr, err := s.auth.Login(model.LoginTypeEmail, req.BasicID, []byte(req.Password))
	return s.respondOnUser(c, authUsr, resp, http.StatusOK, err)
}

func (s *Handler) LoginPhone(c context.Context, req *authms.BasicAuthRequest, resp *authms.Response) error {
	authUsr, err := s.auth.Login(model.LoginTypePhone, req.BasicID, []byte(req.Password))
	return s.respondOnUser(c, authUsr, resp, http.StatusOK, err)
}

func (s *Handler) LoginOAuth(c context.Context, req *authms.OAuthRequest, resp *authms.Response) error {
	authUsr, err := s.auth.Login(model.LoginTypeUsername, req.OAuth.AppToken, nil)
	return s.respondOnUser(c, authUsr, resp, http.StatusOK, err)
}

func (s *Handler) UpdatePhone(c context.Context, req *authms.UpdateRequest, resp *authms.Response) error {
	usr, err := s.auth.UpdateIdentifier(req.Token, model.LoginTypePhone, req.User.Phone.Value)
	return s.respondOnUser(c, usr, resp, http.StatusOK, err)
}

func (s *Handler) UpdateOauth(c context.Context, req *authms.UpdateRequest, resp *authms.Response) error {
	usr, err := s.auth.UpdateIdentifier(req.Token, model.LoginTypeFacebook, req.User.OAuths["facebook"].AppToken)
	return s.respondOnUser(c, usr, resp, http.StatusOK, err)
}

func (s *Handler) VerifyPhone(c context.Context, req *authms.SMSVerificationRequest, resp *authms.SMSVerificationResponse) error {
	vs, err := s.auth.SendVerCode(req.Token, model.LoginTypePhone, req.Phone)
	return s.respondOnSMS(c, vs, req.Phone, resp, err)
}

func (s *Handler) VerifyPhoneCode(c context.Context, req *authms.SMSVerificationCodeRequest, resp *authms.SMSVerificationResponse) error {
	r, err := s.auth.VerifyDBT(model.LoginTypePhone, strconv.FormatInt(req.UserID, 10), []byte(req.Code))
	return s.respondOnSMS(c, nil, r.Address, resp, err)
}

func (s *Handler) respondOnSMS(ctx context.Context, st *model.DBTStatus, addr string, resp *authms.SMSVerificationResponse, err error) error {
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
	var verified bool
	var expiry time.Time
	if st != nil {
		verified = false
		expiry = st.ExpiresAt
	} else {
		verified = true
	}
	resp.Status = &authms.SMSVerificationStatus{
		Phone:     addr,
		Verified:  verified,
		ExpiresAt: expiry.Format(time.RFC3339),
		Token:     "redundant",
	}
	return nil
}

func (s *Handler) respondOnUser(ctx context.Context, usr *model.User, resp *authms.Response, code int32, err error) error {
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
	if usr == nil {
		log := ctx.Value(ctxKeyLog).(*logrus.Entry)
		log.Errorf("Internal error: nil User and error while responding on user")
		resp.Detail = internalErrorMessage
		resp.Code = http.StatusInternalServerError
		return nil
	}
	var usrID int64
	usrID, err = strconv.ParseInt(usr.ID, 10, 64)
	if err != nil {
		log := ctx.Value(ctxKeyLog).(*logrus.Entry)
		log.Errorf("Internal error: userID not a number: %v", err)
		resp.Detail = internalErrorMessage
		resp.Code = http.StatusInternalServerError
		return nil
	}
	var phone *authms.Value
	var email *authms.Value
	var fb = make(map[string]*authms.OAuth)
	if usr.Phone.Address != "" {
		phone = &authms.Value{Value: usr.Phone.Address, Verified: usr.Phone.Verified}
	}
	if usr.Email.Address != "" {
		phone = &authms.Value{Value: usr.Email.Address, Verified: usr.Email.Verified}
	}
	if usr.Facebook.FacebookID != "" {
		fb = map[string]*authms.OAuth{"facebook": {
			AppName:   "facebook",
			AppToken:  usr.Facebook.FacebookToken,
			AppUserID: usr.Facebook.FacebookID,
		},
		}
	}
	resp.Id = s.name
	resp.Code = code
	resp.User = &authms.User{
		ID:       usrID,
		Token:    usr.JWT,
		UserName: usr.UserName.Value,
		Phone:    phone,
		Email:    email,
		OAuths:   fb,
	}
	return nil
}
