package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"os"

	"github.com/sirupsen/logrus"
	testingH "github.com/tomogoma/authms/testing"
	"github.com/tomogoma/go-commons/errors"
)

func init() {
	// TODO test log outputs maybe?
	logrus.SetOutput(os.Stdout)
}

func TestNewHandler(t *testing.T) {
	tt := []struct {
		name   string
		auth   Auth
		guard  Guard
		expErr bool
	}{
		{name: "valid deps", auth: &testingH.AuthenticationMock{}, guard: &testingH.GuardMock{}, expErr: false},
		{name: "nil auth", auth: nil, guard: &testingH.GuardMock{}, expErr: true},
		{name: "nil guard", auth: &testingH.AuthenticationMock{}, guard: nil, expErr: true},
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
		reqBody       string
		reqWBasicAuth bool
		expStatusCode int
		auth          Auth
		guard         Guard
	}{
		// valuse starting and ending with "_" are place holders for variables
		// e.g. _loginType_ is a place holder for "any (valid) login type"
		{
			name:          "register",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/_loginType_/register",
			reqMethod:     http.MethodPut,
			reqBody:       "{}",
			expStatusCode: http.StatusCreated,
		},
		{
			name:          "register guard error",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{ExpAPIKValidErr: errors.Newf("guard error")},
			reqURLSuffix:  "/_loginType_/register",
			reqMethod:     http.MethodPut,
			reqBody:       "{}",
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "register bad body",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/_loginType_/register",
			reqMethod:     http.MethodPut,
			reqBody:       "{bad json]",
			expStatusCode: http.StatusBadRequest,
		},
		{
			name:          "register self error",
			auth:          &testingH.AuthenticationMock{ExpRegSelfErr: errors.Newf("auth reg self error")},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/_loginType_/register?selfReg=true",
			reqMethod:     http.MethodPut,
			reqBody:       "{}",
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "register self (lock phone) error",
			auth:          &testingH.AuthenticationMock{ExpRegSelfBLPErr: errors.Newf("auth reg self by locked phone error")},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/_loginType_/register?selfReg=device",
			reqMethod:     http.MethodPut,
			reqBody:       "{}",
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "register other error",
			auth:          &testingH.AuthenticationMock{ExpRegOtherErr: errors.Newf("auth reg other error")},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/_loginType_/register",
			reqMethod:     http.MethodPut,
			reqBody:       "{}",
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "login",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/_loginType_/login",
			reqMethod:     http.MethodPost,
			reqWBasicAuth: true,
			reqBody:       "{}",
			expStatusCode: http.StatusOK,
		},
		{
			name:          "login guard error",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{ExpAPIKValidErr: errors.Newf("guard error")},
			reqURLSuffix:  "/_loginType_/login",
			reqMethod:     http.MethodPost,
			reqWBasicAuth: true,
			reqBody:       "{}",
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "login no basic auth",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/_loginType_/login",
			reqMethod:     http.MethodPost,
			reqWBasicAuth: false,
			reqBody:       "{}",
			expStatusCode: http.StatusUnauthorized,
		},
		{
			name:          "login error",
			auth:          &testingH.AuthenticationMock{ExpLoginErr: errors.Newf("auth login error")},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/_loginType_/login",
			reqMethod:     http.MethodPost,
			reqWBasicAuth: true,
			reqBody:       "{}",
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "update",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/_loginType_/update",
			reqMethod:     http.MethodPost,
			reqBody:       "{}",
			expStatusCode: http.StatusOK,
		},
		{
			name:          "update guard error",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{ExpAPIKValidErr: errors.Newf("guard error")},
			reqURLSuffix:  "/_loginType_/update",
			reqMethod:     http.MethodPost,
			reqBody:       "{}",
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "update bad body",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/_loginType_/update",
			reqMethod:     http.MethodPost,
			reqBody:       "{bad json]",
			expStatusCode: http.StatusBadRequest,
		},
		{
			name:          "update auth error",
			auth:          &testingH.AuthenticationMock{ExpUpdIDerErr: errors.Newf("auth update identifier error")},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/_loginType_/update",
			reqMethod:     http.MethodPost,
			reqBody:       "{}",
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "send ver code",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/_loginType_/verify/address",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusOK,
		},
		{
			name:          "send ver code guard error",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{ExpAPIKValidErr: errors.Newf("guard error")},
			reqURLSuffix:  "/_loginType_/verify/address",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "send ver code auth error",
			auth:          &testingH.AuthenticationMock{ExpSndVerCodeErr: errors.Newf("auth send ver code error")},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/_loginType_/verify/address",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "verify code",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/users/_userID_/verify/_dbToken_",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusOK,
		},
		{
			name:          "verify code guard error",
			auth:          &testingH.AuthenticationMock{},
			guard:         &testingH.GuardMock{ExpAPIKValidErr: errors.Newf("guard error")},
			reqURLSuffix:  "/users/_userID_/verify/_dbToken_",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "verify code ver dbt err",
			auth:          &testingH.AuthenticationMock{ExpVerDBTErr: errors.Newf("auth ver DBT error")},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/users/_userID_/verify/_dbToken_",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusInternalServerError,
		},
		{
			name:          "verify code ver and extend dbt err",
			auth:          &testingH.AuthenticationMock{ExpVerExtDBTErr: errors.Newf("auth ver and extend DBT error")},
			guard:         &testingH.GuardMock{ExpAPIKValidUsrID: "12345"},
			reqURLSuffix:  "/users/_userID_/verify/_dbToken_?extend=true",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusInternalServerError,
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
				bytes.NewReader([]byte(tc.reqBody)),
			)
			if err != nil {
				t.Fatalf("Error setting up: new request: %v", err)
			}
			if tc.reqWBasicAuth {
				req.SetBasicAuth("username", "password")
			}

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
