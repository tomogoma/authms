package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tomogoma/authms/model"
	testingH "github.com/tomogoma/authms/testing"
	"github.com/tomogoma/go-commons/errors"
)

type AuthMock struct {
	errors.NotImplErrCheck
	errors.AuthErrCheck
	errors.ClErrCheck
	expUser       *model.User
	expErr        error
	expVerLogin   *model.VerifLogin
	expObfuscAddr string
}

func (a *AuthMock) RegisterSelf(loginType, userType, id string, secret []byte) (*model.User, error) {
	return a.expUser, a.expErr
}
func (a *AuthMock) RegisterSelfByLockedPhone(userType, devID, number string, secret []byte) (*model.User, error) {
	return a.expUser, a.expErr
}
func (a *AuthMock) RegisterOther(JWT, newLoginType, userType, id, groupID string) (*model.User, error) {
	return a.expUser, a.expErr
}
func (a *AuthMock) UpdateIdentifier(JWT, loginType, newId string) (*model.User, error) {
	return a.expUser, a.expErr
}
func (a *AuthMock) UpdatePassword(JWT string, old, newPass []byte) error {
	return a.expErr
}
func (a *AuthMock) SetPassword(loginType, userID string, dbt, pass []byte) (*model.VerifLogin, error) {
	return a.expVerLogin, a.expErr
}
func (a *AuthMock) SendVerCode(JWT, loginType, toAddr string) (string, error) {
	return a.expObfuscAddr, a.expErr
}
func (a *AuthMock) SendPassResetCode(loginType, toAddr string) (string, error) {
	return a.expObfuscAddr, a.expErr
}
func (a *AuthMock) VerifyAndExtendDBT(lt, usrID string, dbt []byte) (string, error) {
	return a.expObfuscAddr, a.expErr
}
func (a *AuthMock) VerifyDBT(loginType, userID string, dbt []byte) (*model.VerifLogin, error) {
	return a.expVerLogin, a.expErr
}
func (a *AuthMock) Login(loginType, identifier string, password []byte) (*model.User, error) {
	return a.expUser, a.expErr
}

func TestNewHandler(t *testing.T) {
	tt := []struct {
		name   string
		auth   Auth
		guard  Guard
		expErr bool
	}{
		{name: "valid deps", auth: &AuthMock{}, guard: &testingH.GuardMock{}, expErr: false},
		{name: "nil auth", auth: nil, expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewHandler(tc.auth, tc.guard)
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

func TestHandler_handleRoute(t *testing.T) {
	tt := []struct {
		name          string
		reqURLSuffix  string
		reqMethod     string
		expStatusCode int
		auth          Auth
		guard         Guard
	}{
		{
			name:          "register",
			auth:          &AuthMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/loginType/register",
			reqMethod:     http.MethodPut,
			expStatusCode: http.StatusCreated,
		},
		{
			name:          "login",
			auth:          &AuthMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/login-type/login",
			reqMethod:     http.MethodPost,
			expStatusCode: http.StatusOK,
		},
		{
			name:          "send ver code",
			auth:          &AuthMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/login-type/verify/address",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusOK,
		},
		{
			name:          "verify code",
			auth:          &AuthMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/users/userid/verify/a-dbt-token",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusOK,
		},
		{
			name:          "update",
			auth:          &AuthMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/login-type/update",
			reqMethod:     http.MethodPost,
			expStatusCode: http.StatusOK,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(t, tc.auth, tc.guard)
			srvr := httptest.NewServer(h)
			defer srvr.Close()
			req, err := http.NewRequest(
				tc.reqMethod,
				srvr.URL+tc.reqURLSuffix,
				bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				t.Fatalf("Error setting up, new request: %v", err)
			}
			req.SetBasicAuth("uname", "password")
			cl := &http.Client{}
			resp, err := cl.Do(req)
			if err != nil {
				t.Fatalf("Do request error: %v", err)
			}
			if resp.StatusCode != tc.expStatusCode {
				t.Errorf("Expected status code %d, got %s",
					tc.expStatusCode, resp.Status)
			}
		})
	}
}

func newHandler(t *testing.T, a Auth, g Guard) http.Handler {
	h, err := NewHandler(a, g)
	if err != nil {
		t.Fatalf("http.NewHandler(): %v", err)
	}
	return h
}
