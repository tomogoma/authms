package model_test

import (
	"flag"
	"runtime"
	"testing"
	"time"

	"reflect"

	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/facebook"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/auth/token"
	configH "github.com/tomogoma/go-commons/config"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/go-commons/errors"
	"golang.org/x/crypto/bcrypt"
)

type TokenConfigMock struct {
	TknKyFile string `yaml:"tokenKeyFile,omitempty"`
}

func (c TokenConfigMock) KeyFile() string {
	return c.TknKyFile
}

type ConfigMock struct {
	Database cockroach.DSN   `json:"database,omitempty"`
	Token    TokenConfigMock `json:"token,omitempty"`
}

type OAuthHandlerMock struct {
	errors.AuthErrCheck
	ExpValTknClld bool
	ExpValid      bool
	AppUserID     string
	ExpErr        error
	ValTknClld    bool
}

func (oa *OAuthHandlerMock) ValidateToken(token string) (model.OAuthResponse, error) {
	oa.ValTknClld = true
	return &facebook.Response{
		ResponseData: facebook.ResponseData{
			Valid: oa.ExpValid,
			UsrID: oa.AppUserID,
		},
	}, oa.ExpErr
}

type SMSerMock struct {
	errors.NotImplErrCheck
	ExpErr error
}

func (s *SMSerMock) SMS(toPhone, message string) error {
	return s.ExpErr
}

type DBHelperMock struct {
	ExpErr                     error
	ExpUser                    *authms.User
	ExpHist                    []*authms.History
	ExpUserID                  int64
	ExpIsNotFound              bool
	SaveUserCalled             bool
	SaveHistoryCalled          bool
	GetUserCalled              bool
	UpdatePhoneCalled          bool
	UpdatePassCalled           bool
	UpdateAppUserIDCalled      bool
	T                          *testing.T
	IgnoreDefaultVerifiedCheck bool
	ExpIsDuplicate             bool
	ExpLoginVer                model.LoginVerification
	ExpLoginVers               []model.LoginVerification
}

func (d *DBHelperMock) SaveUser(u *authms.User) error {
	for _, oauth := range u.OAuths {
		if oauth.Verified == false {
			d.T.Error("expected OAuth verified=true but got false")
		}
	}
	if u.Phone != nil && u.Phone.Verified == true {
		d.T.Error("expected Phone verified=false but got true")
	}
	if u.Email != nil && u.Email.Verified == true {
		d.T.Error("expected Email verified=false but got true")
	}
	d.SaveUserCalled = true
	return d.ExpErr
}
func (d *DBHelperMock) GetByUserName(uname, pass string) (*authms.User, error) {
	d.GetUserCalled = true
	return d.ExpUser, d.ExpErr
}
func (d *DBHelperMock) GetByEmail(email, pass string) (*authms.User, error) {
	d.GetUserCalled = true
	return d.ExpUser, d.ExpErr
}
func (d *DBHelperMock) GetByPhone(phone, pass string) (*authms.User, error) {
	d.GetUserCalled = true
	return d.ExpUser, d.ExpErr
}
func (d *DBHelperMock) GetByAppUserID(appName, appUserID string) (*authms.User, error) {
	d.GetUserCalled = true
	return d.ExpUser, d.ExpErr
}
func (d *DBHelperMock) UpsertLoginVerification(code model.LoginVerification) (model.LoginVerification, error) {
	return d.ExpLoginVer, d.ExpErr
}
func (d *DBHelperMock) GetLoginVerifications(verificationType string, userID, offset, count int64) ([]model.LoginVerification, error) {
	return d.ExpLoginVers, d.ExpErr
}

func (d *DBHelperMock) GetHistory(userID int64, offset, count int, accessType ...string) ([]*authms.History, error) {
	return d.ExpHist, d.ExpErr
}
func (d *DBHelperMock) SaveHistory(*authms.History) error {
	d.SaveHistoryCalled = true
	return d.ExpErr
}
func (d *DBHelperMock) IsNotFoundError(err error) bool {
	return d.ExpIsNotFound
}
func (d *DBHelperMock) IsDuplicateError(err error) bool {
	return d.ExpIsDuplicate
}

func (d *DBHelperMock) UpdatePhone(userID int64, newPhone *authms.Value) error {
	d.UpdatePhoneCalled = true
	if !d.IgnoreDefaultVerifiedCheck && newPhone.Verified == true {
		d.T.Error("expected Phone verified=false but got true")
	}
	return d.ExpErr
}

func (d *DBHelperMock) UpdatePassword(userID int64, oldPass, newPassword string) error {
	d.UpdatePassCalled = true
	return d.ExpErr
}

func (d *DBHelperMock) UpdateAppUserID(userID int64, new *authms.OAuth) error {
	d.UpdateAppUserIDCalled = true
	if new.Verified == false {
		d.T.Error("expected OAuth verified=true but got false")
	}
	return d.ExpErr
}

func (d *DBHelperMock) UserExists(u *authms.User) (int64, error) {
	return d.ExpUserID, d.ExpErr
}

type RegisterTestCase struct {
	Desc      string
	User      *authms.User
	OAHandler *OAuthHandlerMock
	DBHelper  *DBHelperMock
	ExpErr    bool
	DevID     string
	Token     string
}

var conf = &ConfigMock{}
var tokenGen *token.JWTHandler
var confFile = flag.String(
	"conf",
	config.DefaultConfPath,
	"/path/to/config/file.conf.yml",
)

func init() {
	flag.Parse()
}

func TestNew(t *testing.T) {
	setUp(t)
	newAuth(t, &DBHelperMock{T: t}, &OAuthHandlerMock{}, &SMSerMock{})
}

func TestAuth_Register(t *testing.T) {
	appName := "facebook"
	cases := []RegisterTestCase{
		{
			Desc:      "Valid args",
			User:      &authms.User{ID: 123, UserName: "some-name", Password: "some-password"},
			OAHandler: &OAuthHandlerMock{},
			DBHelper:  &DBHelperMock{T: t},
			ExpErr:    false,
			DevID:     "test-dev",
		},
		{
			Desc: "Valid args, verified=true",
			User: &authms.User{
				ID: 123,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName:   appName,
						AppUserID: "1",
						Verified:  true,
						AppToken:  "some-app-token",
					},
				},
				Phone: &authms.Value{Verified: true},
				Email: &authms.Value{Verified: true},
			},
			OAHandler: &OAuthHandlerMock{AppUserID: "1", ExpValid: true},
			DBHelper:  &DBHelperMock{T: t},
			ExpErr:    false,
			DevID:     "test-dev",
		},
		{
			Desc:      "Invalid user (has password only)",
			User:      &authms.User{Password: "exp-password-deleted"},
			OAHandler: &OAuthHandlerMock{},
			DBHelper:  &DBHelperMock{T: t},
			ExpErr:    true,
			DevID:     "test-dev",
		},
		{
			Desc: "Nil user",
			User: nil, OAHandler: &OAuthHandlerMock{},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    "test-dev"},
		{
			Desc: "OAuth user",
			User: &authms.User{
				ID: 123,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName:   appName,
						AppUserID: "test-oauth-app-uid",
						AppToken:  "some-test-app-token",
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpValTknClld: true,
				ExpValid:      true,
				AppUserID:     "test-oauth-app-uid",
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   false,
			DevID:    "test-dev",
		},
		{
			Desc:      "Missing DevID",
			User:      &authms.User{ID: 123, UserName: "some-name", Password: "some-password"},
			OAHandler: &OAuthHandlerMock{},
			DBHelper:  &DBHelperMock{T: t},
			ExpErr:    true,
		},
	}
	for _, c := range cases {
		func() {
			setUp(t)
			a := newAuth(t, c.DBHelper, c.OAHandler, &SMSerMock{})
			err := a.Register(c.User, c.DevID, "")
			if c.User != nil && c.User.Password != "" {
				t.Errorf("%s - expected password to be cleared"+
					" but was not", c.Desc)
			}
			runtime.Gosched()
			if c.OAHandler.ExpValTknClld && !c.OAHandler.ValTknClld {
				t.Errorf("%s - validate oath token not called", c.Desc)
			}
			if c.ExpErr {
				if err == nil {
					t.Errorf("%s - Expected an error but"+
						" got nil", c.Desc)
				}
				return
			} else if err != nil {
				t.Errorf("%s: auth.Register(): %v", c.Desc, err)
				return
			}
			if !c.DBHelper.SaveUserCalled {
				t.Errorf("%s - save user not called", c.Desc)
			}
			if !c.DBHelper.SaveHistoryCalled {
				t.Errorf("%s - Save history not called", c.Desc)
			}
		}()
	}
}

func genToken(t *testing.T, usrID int64, devID string) string {
	clm := model.NewClaim(usrID, devID, 1*time.Hour)
	tkn, err := tokenGen.Generate(clm)
	if err != nil {
		t.Fatalf("Error setting up: token.Generate(): %v", err)
	}
	return tkn
}

func TestAuth_UpdatePhone(t *testing.T) {
	usrID := int64(123)
	devID := "test-dev"
	tkn := genToken(t, usrID, devID)
	cases := []RegisterTestCase{
		{
			Desc: "Valid phone",
			User: &authms.User{
				ID:    usrID,
				Phone: &authms.Value{Value: "+254123456789"},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   false,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "Valid phone and verified=true (exp verified=false sent to dbhelper)",
			User: &authms.User{
				ID: usrID,
				Phone: &authms.Value{
					Value:    "+254123456789",
					Verified: true},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   false,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "Invalid phone",
			User: &authms.User{
				ID:    usrID,
				Phone: &authms.Value{},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "nil phone",
			User: &authms.User{
				ID: usrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc:     "nil user",
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "no userID",
			User: &authms.User{
				Phone: &authms.Value{Value: "+254123456789"},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "Invalid token",
			User: &authms.User{
				ID:    usrID + 5,
				Phone: &authms.Value{Value: "+254123456789"},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
	}
	for _, c := range cases {
		func() {
			setUp(t)
			a := newAuth(t, c.DBHelper, c.OAHandler, &SMSerMock{})
			err := a.UpdatePhone(c.User, c.Token, c.DevID, "")
			runtime.Gosched()
			if c.ExpErr {
				if err == nil {
					t.Errorf("%s - Expected an error but"+
						" got nil", c.Desc)
				}
				return
			} else if err != nil {
				t.Errorf("%s: auth.UpdatePhone(): %v", c.Desc, err)
				return
			}
			if !c.DBHelper.UpdatePhoneCalled {
				t.Errorf("%s - save user not called", c.Desc)
			}
			if !c.DBHelper.SaveHistoryCalled {
				t.Errorf("%s - Save history not called", c.Desc)
			}
		}()
	}
}

func TestAuth_UpdateOAuth(t *testing.T) {
	usrID := int64(123)
	devID := "test-dev"
	tkn := genToken(t, usrID, devID)
	appName := "facebook"
	appUsrID := "test-app-userID"
	cases := []RegisterTestCase{
		{
			Desc: "Valid OAuth",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName:   appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr:        nil,
				ExpValid:      true,
				AppUserID:     appUsrID,
				ExpValTknClld: true,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   false,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "Mismatched OAuth names",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					"mismatched-name": {
						AppName:   "mismatched-name",
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpValTknClld: false,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "missing app user ID",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					appName: {AppName: appName},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr:        nil,
				ExpValid:      true,
				AppUserID:     appUsrID,
				ExpValTknClld: true,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "Missing user ID",
			User: &authms.User{
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName:   appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr:    nil,
				ExpValid:  true,
				AppUserID: appUsrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "Nil OAuth",
			User: &authms.User{
				ID: usrID,
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr:    nil,
				ExpValid:  true,
				AppUserID: appUsrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "Valid OAuth and verified=true (exp dbhelper to be passed to false)",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName:   appName,
						AppUserID: appUsrID,
						Verified:  true,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr:        nil,
				ExpValid:      true,
				AppUserID:     appUsrID,
				ExpValTknClld: true,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   false,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "Invalid token",
			User: &authms.User{
				ID: usrID + 5,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName:   appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr:    nil,
				ExpValid:  true,
				AppUserID: appUsrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "OAuth report error",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName:   appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr:        errors.New(""),
				ExpValid:      true,
				AppUserID:     appUsrID,
				ExpValTknClld: true,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    devID,
			Token:    tkn,
		},
		{
			Desc: "Missing device ID",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName:   appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr:    nil,
				ExpValid:  true,
				AppUserID: appUsrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   true,
			DevID:    "",
			Token:    tkn,
		},
	}
	for _, c := range cases {
		func() {
			setUp(t)
			a := newAuth(t, c.DBHelper, c.OAHandler, &SMSerMock{})
			err := a.UpdateOAuth(c.User, appName, c.Token, c.DevID, "")
			runtime.Gosched()
			if c.OAHandler.ExpValTknClld && !c.OAHandler.ValTknClld {
				t.Errorf("%s - validate token was not called", c.Desc)
			}
			if c.ExpErr {
				if err == nil {
					t.Errorf("%s - Expected an error but"+
						" got nil", c.Desc)
				}
				return
			} else if err != nil {
				t.Errorf("%s: auth.UpdatePhone(): %v", c.Desc, err)
				return
			}
			if !c.DBHelper.UpdateAppUserIDCalled {
				t.Errorf("%s - update userID not called", c.Desc)
			}
			if !c.DBHelper.SaveHistoryCalled {
				t.Errorf("%s - Save history not called", c.Desc)
			}
		}()
	}
}

func TestAuth_VerifyPhone(t *testing.T) {
	type VerifyTestCase struct {
		ExpErr     bool
		ExpNotImpl bool
		Desc       string
		Req        *authms.SMSVerificationRequest
		SMS        model.SMSer
		ExpStatus  *authms.SMSVerificationStatus
		DB         *DBHelperMock
	}
	phone := "+254712345678"
	devID := "test-dev"
	userID := int64(1234)
	validToken := genToken(t, userID, devID)
	tcs := []VerifyTestCase{
		{
			Desc:   "Successful validation",
			ExpErr: false,
			Req: &authms.SMSVerificationRequest{
				DeviceID: devID,
				UserID:   userID,
				Token:    validToken,
				Phone:    phone,
			},
			SMS: &SMSerMock{
				ExpErr: nil,
			},
			ExpStatus: &authms.SMSVerificationStatus{
				Phone:    phone,
				Token:    validToken,
				Verified: false,
			},
			DB: &DBHelperMock{T: t},
		},
		{
			Desc:   "Nil SMSer",
			ExpErr: true, ExpNotImpl: true,
			Req: &authms.SMSVerificationRequest{
				DeviceID: devID,
				UserID:   userID,
				Token:    validToken,
				Phone:    phone,
			},
			SMS: nil,
			DB:  &DBHelperMock{T: t},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.Desc, func(t *testing.T) {
			setUp(t)
			a := newAuth(t, tc.DB, &OAuthHandlerMock{}, tc.SMS)
			r, err := a.VerifyPhone(tc.Req, "")
			runtime.Gosched()
			if !tc.DB.SaveHistoryCalled {
				t.Errorf("%s - save history was not called", tc.Desc)
			}
			if tc.ExpErr {
				if err == nil {
					t.Fatalf("%s - expected error but "+
						"got nil", tc.Desc)
				}
				if tc.ExpNotImpl != a.IsNotImplementedError(err) {
					t.Errorf("%s - expected IsNotImplementedError() %t, got %t",
						tc.Desc, tc.ExpNotImpl, a.IsNotImplementedError(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("%s - auth.VerifyPhone(): %v", tc.Desc, err)
				return
			}
			if r.ExpiresAt == "" {
				t.Errorf("Expires at was empty")
			}
			r.ExpiresAt = ""
			if !reflect.DeepEqual(r, tc.ExpStatus) {
				t.Errorf("Status mismatch\ngot:\t%+v\nexpect:\t%+v",
					r, tc.ExpStatus)
				return
			}
		})
	}
}

func TestAuth_VerifyPhoneCode(t *testing.T) {
	type VerifyTestCase struct {
		ExpErr    bool
		Desc      string
		Req       *authms.SMSVerificationCodeRequest
		SMS       *SMSerMock
		DB        *DBHelperMock
		ExpStatus *authms.SMSVerificationStatus
	}
	phone := "+254712345678"
	devID := "test-dev"
	userID := int64(1234)
	validToken := genToken(t, userID, devID)
	code := "1876"
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Error setting up: hash SMS Code")
	}
	tcs := []VerifyTestCase{
		{
			Desc:   "Successful validation",
			ExpErr: false,
			Req: &authms.SMSVerificationCodeRequest{
				DeviceID: devID,
				UserID:   userID,
				Code:     code,
				Token:    validToken,
			},
			SMS: &SMSerMock{
				ExpErr: nil,
			},
			ExpStatus: &authms.SMSVerificationStatus{
				Verified: true,
				Phone:    phone,
			},
			DB: &DBHelperMock{
				T: t,
				IgnoreDefaultVerifiedCheck: true,
				ExpLoginVers: []model.LoginVerification{
					{
						SubjectValue: phone,
						IsUsed:       false,
						Expiry:       time.Now().Add(1 * time.Minute),
						CodeHash:     []byte{},
					},
					{
						SubjectValue: phone,
						IsUsed:       false,
						Expiry:       time.Now().Add(1 * time.Minute),
						CodeHash:     codeHash,
					},
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.Desc, func(t *testing.T) {
			setUp(t)
			a := newAuth(t, tc.DB, &OAuthHandlerMock{}, tc.SMS)
			r, err := a.VerifyPhoneCode(tc.Req, "")
			runtime.Gosched()
			if !tc.DB.SaveHistoryCalled {
				t.Errorf("%s - save history was not called", tc.Desc)
			}
			if tc.ExpErr {
				if err == nil {
					t.Errorf("%s - expected error but "+
						"got nil", tc.Desc)
				}
				return
			}
			if err != nil {
				t.Errorf("%s - auth.VerifyPhone(): %v", tc.Desc, err)
				return
			}
			if !tc.DB.UpdatePhoneCalled {
				t.Errorf("%s - update phone was not called", tc.Desc)
			}
			if !reflect.DeepEqual(r, tc.ExpStatus) {
				t.Errorf("Status mismtach\nGot:\t%+v\nExpect:\t%+v",
					r, tc.ExpStatus)
				return
			}
		})
	}
}

func TestAuth_UpdatePassword(t *testing.T) {
	usrID := int64(123)
	devID := "test-dev"
	cases := []RegisterTestCase{
		{
			Desc: "Valid password",
			User: &authms.User{
				ID:       usrID,
				Password: "some-new-password",
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr:   false,
			DevID:    devID,
		},
	}
	for _, c := range cases {
		func() {
			setUp(t)
			a := newAuth(t, c.DBHelper, c.OAHandler, &SMSerMock{})
			err := a.UpdatePassword(c.User.ID, "some-old-pass",
				c.User.Password, c.DevID, "")
			runtime.Gosched()
			if c.ExpErr {
				if err == nil {
					t.Errorf("%s - Expected an error but"+
						" got nil", c.Desc)
				}
				return
			} else if err != nil {
				t.Errorf("%s: auth.UpdatePassword(): %v", c.Desc, err)
				return
			}
			if !c.DBHelper.UpdatePassCalled {
				t.Errorf("%s - update password was not called", c.Desc)
			}
			if !c.DBHelper.SaveHistoryCalled {
				t.Errorf("%s - Save history not called", c.Desc)
			}
		}()
	}
}

func TestAuth_LoginOAuth(t *testing.T) {
	type LoginOAuthTestCase struct {
		ExpErr    bool
		Desc      string
		OAHandler *OAuthHandlerMock
		DBHelper  *DBHelperMock
		OAuth     *authms.OAuth
		DevID     string
	}
	cases := []LoginOAuthTestCase{
		{Desc: "Valid Creds", ExpErr: false,
			DBHelper: &DBHelperMock{ExpUser: &authms.User{ID: 123}},
			OAHandler: &OAuthHandlerMock{ExpValTknClld: true,
				AppUserID: "test-app-user-id", ExpValid: true},
			OAuth: &authms.OAuth{AppName: "facebook", AppUserID: "test-app-user-id"}, DevID: "Tes-devID"},
		{Desc: "Invalid OAuth Creds", ExpErr: true,
			DBHelper:  &DBHelperMock{T: t},
			OAHandler: &OAuthHandlerMock{ExpValTknClld: true, ExpErr: errors.New("")},
			OAuth:     &authms.OAuth{AppName: "facebook"}, DevID: "Tes-devID"},
		{Desc: "Nil OAuth", ExpErr: true,
			DBHelper:  &DBHelperMock{T: t},
			OAHandler: &OAuthHandlerMock{ExpValTknClld: false, ExpErr: errors.New("")},
			OAuth:     nil, DevID: "Tes-devID"},
	}
	for _, c := range cases {
		func() {
			a := newAuth(t, c.DBHelper, c.OAHandler, &SMSerMock{})
			usr, err := a.LoginOAuth(c.OAuth, c.DevID, "")
			runtime.Gosched()
			if c.OAHandler.ExpValTknClld && !c.OAHandler.ValTknClld {
				t.Errorf("%s - validate oath token not called", c.Desc)
			}
			if c.ExpErr {
				if err == nil {
					t.Errorf("%s - expected error But got nil",
						c.Desc)
				}
				return
			} else if err != nil {
				t.Errorf("%s - auth.loginOAuth(): %s", c.Desc, err)
				return
			}
			if !c.DBHelper.SaveHistoryCalled {
				t.Errorf("%s - Save history not called", c.Desc)
			}
			if usr != c.DBHelper.ExpUser {
				t.Errorf("%s - user %+v was not expected %+v",
					c.Desc, usr, c.DBHelper.ExpUser)
			}
			if usr.Token == "" {
				t.Errorf("%s - missing token for user", c.Desc)
			}
			if !c.DBHelper.GetUserCalled {
				t.Errorf("%s - get user not called", c.Desc)
			}
		}()
	}
}

func newAuth(t *testing.T, db *DBHelperMock, oa *OAuthHandlerMock, smser model.SMSer) *model.Auth {
	var err error
	tokenGen, err = token.NewJWTHandler(conf.Token)
	if err != nil {
		t.Fatalf("token.NewGenerator(): %v", err)
	}
	a, err := model.New(tokenGen, db, model.WithSMSer(smser), model.WithFB(oa))
	if err != nil {
		t.Fatalf("auth.New(): %v", err)
	}
	return a
}

func setUp(t *testing.T) {
	if err := configH.ReadYamlConfig(*confFile, conf); err != nil {
		t.Fatal(err)
	}
}

//func TestAuth_LoginOAuth2(t *testing.T) {
//	// TODO p>m login/token attempts from an ip address over duration t
//	t.Fatal("untested:\nCheck:\n3 failed attempts for a user from an IP in an hour blocked\n6 attempts from a user in an hour blocked")
//}
//
//func TestAPIKeysEnforced(t *testing.T) {
//	t.Fatal("untested:\nAccess auth services only if client (microservice) / API key combo is recogonized")
//}
//
//func TestAuth_RegisterUser_limitsFailure(t *testing.T) {
//	//
//	//a := newAuth(histM, t)
//	//usr := &testhelper.User{}
//	//a.RegisterUser(usr, "pass", "127.0.0.1", "test", "authms")
//	//a.RegisterUser(usr, "pass", "127.0.0.1", "test", "authms")
//	//a.RegisterUser(usr, "pass", "127.0.0.1", "test", "authms")
//	//actUsr := testhelper.User{UName: "uname", Password: "pass"}
//	//svdUsr, err := a.RegisterUser(actUsr, "pass", "127.0.0.1", "test", "authms")
//	//if err == nil || err !=
//	// TODO p>m login/token attempts from an ip address over duration t
//	// 1. Check db if failed attempts of each type login, reg, token were made in
//	//    the past configured max time
//	// 2. check, for each type, if the failed count exceeds the configured max
//	// 3. if it does, return specific error
//	t.Fatal("untested:\nCheck:\n3 failed attempts for a user from an IP in an hour blocked\n6 attempts from a user in an hour blocked")
//}
