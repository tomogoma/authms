package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/errors"
)

type AuthMock struct {
	errors.AuthErrCheck
	errors.ClErrCheck
	expUser         *authms.User
	expSMSVerStatus *authms.SMSVerificationStatus
	expErr          error
}

func (a *AuthMock) Register(user *authms.User, devID, rIP string) error {
	return a.expErr
}
func (a *AuthMock) LoginUserName(uName, pass, devID, rIP string) (*authms.User, error) {
	return a.expUser, a.expErr
}
func (a *AuthMock) LoginEmail(email, pass, devID, rIP string) (*authms.User, error) {
	return a.expUser, a.expErr
}
func (a *AuthMock) LoginPhone(phone, pass, devID, rIP string) (*authms.User, error) {
	return a.expUser, a.expErr
}
func (a *AuthMock) LoginOAuth(app *authms.OAuth, devID, rIP string) (*authms.User, error) {
	return a.expUser, a.expErr
}
func (a *AuthMock) UpdatePhone(user *authms.User, token, devID, rIP string) error {
	return a.expErr
}
func (a *AuthMock) UpdateOAuth(user *authms.User, appName, token, devID, rIP string) error {
	return a.expErr
}
func (a *AuthMock) VerifyPhone(req *authms.SMSVerificationRequest, rIP string) (*authms.SMSVerificationStatus, error) {
	return a.expSMSVerStatus, a.expErr
}
func (a *AuthMock) VerifyPhoneCode(req *authms.SMSVerificationCodeRequest, rIP string) (*authms.SMSVerificationStatus, error) {
	return a.expSMSVerStatus, a.expErr
}

func TestNewHandler(t *testing.T) {
	tt := []struct {
		name   string
		auth   Auth
		expErr bool
	}{
		{name: "valid deps", auth: &AuthMock{}, expErr: false},
		{name: "nil auth", auth: nil, expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewHandler(tc.auth)
			if tc.expErr {
				if err == nil {
					t.Fatal("Expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if h == nil {
				t.Fatalf("http.NewHandler() yielded a nil handler!")
			}
		})
	}
}

func TestHandler_HandleRoute(t *testing.T) {
	tt := []struct {
		name           string
		r              *mux.Router
		reqURLSuffix   string
		reqContentType string
		reqBody        []byte
		expErr         bool
		expStatusCode  int
	}{
		{
			name:   "nil router",
			r:      nil,
			expErr: true,
		},
		{
			name:           "register",
			r:              mux.NewRouter(),
			reqURLSuffix:   "/register",
			reqContentType: "application/json",
			reqBody:        []byte("{}"),
			expErr:         false,
			expStatusCode:  http.StatusCreated,
		}, {
			name:           "oauth login",
			r:              mux.NewRouter(),
			reqURLSuffix:   "/oauths/login",
			reqContentType: "application/json",
			reqBody:        []byte("{}"),
			expErr:         false,
			expStatusCode:  http.StatusOK,
		}, {
			name:           "verify phone",
			r:              mux.NewRouter(),
			reqURLSuffix:   "/phones/verify",
			reqContentType: "application/json",
			reqBody:        []byte("{}"),
			expErr:         false,
			expStatusCode:  http.StatusOK,
		}, {
			name:           "verify code",
			r:              mux.NewRouter(),
			reqURLSuffix:   "/codes/verify",
			reqContentType: "application/json",
			reqBody:        []byte("{}"),
			expErr:         false,
			expStatusCode:  http.StatusOK,
		}, {
			name:           "generic login",
			r:              mux.NewRouter(),
			reqURLSuffix:   "/emails/login",
			reqContentType: "application/json",
			reqBody:        []byte("{}"),
			expErr:         false,
			expStatusCode:  http.StatusOK,
		}, {
			name:           "update",
			r:              mux.NewRouter(),
			reqURLSuffix:   "/phones/update",
			reqContentType: "application/json",
			reqBody:        []byte("{}"),
			expErr:         false,
			expStatusCode:  http.StatusOK,
		},
	}
	h := newHandler(t, &AuthMock{})
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := h.HandleRoute(tc.r)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("http.Handler#HandleRoute(): %v", err)
			}
			srvr := httptest.NewServer(tc.r)
			defer srvr.Close()
			resp, err := http.Post(srvr.URL+tc.reqURLSuffix, tc.reqContentType, bytes.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("Unable to make post request: %v", err)
			}
			if resp.StatusCode != tc.expStatusCode {
				t.Errorf("Expected status code %d, got %s",
					tc.expStatusCode, resp.Status)
			}
		})
	}
}

func TestHandler_handleRegistration(t *testing.T) {
	tt := []struct {
		name           string
		auth           *AuthMock
		r              *mux.Router
		reqURLSuffix   string
		reqContentType string
		reqBody        []byte
		expErr         bool
		expStatusCode  int
	}{}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(t, tc.auth)
			err := h.HandleRoute(tc.r)
			if err != nil {
				t.Fatalf("http.Handler#HandleRoute(): %v", err)
			}
			srvr := httptest.NewServer(tc.r)
			defer srvr.Close()
			resp, err := http.Post(srvr.URL+tc.reqURLSuffix, tc.reqContentType, bytes.NewReader(tc.reqBody))
			if err != nil {
				t.Fatalf("Unable to make post request: %v", err)
			}
			if resp.StatusCode != tc.expStatusCode {
				t.Errorf("Expected status code %d, got %s",
					tc.expStatusCode, resp.Status)
			}
		})
	}
}

func newHandler(t *testing.T, a Auth) *Handler {
	h, err := NewHandler(a)
	if err != nil {
		t.Fatalf("http.NewHandler(): %v", err)
	}
	return h
}
