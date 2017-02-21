package auth_test

import (
	"testing"

	"github.com/limetext/log4go"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/go-commons/config"
	"flag"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/authms/auth/oauth/response"
	"github.com/tomogoma/authms/auth/oauth/facebook"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/errors"
	"runtime"
)

type TokenConfigMock struct {
	TknKyFile string `yaml:"tokenKeyFile,omitempty"`
}

func (c TokenConfigMock) TokenKeyFile() string {
	return c.TknKyFile
}

type ConfigMock struct {
	Database cockroach.DSN    `json:"database,omitempty"`
	Token    TokenConfigMock  `json:"token,omitempty"`
}

type OAuthHandlerMock struct {
	ExpValTknClld bool
	ExpValid      bool
	AppUserID     string
	ExpErr        error
	ValTknClld    bool
}

func (oa *OAuthHandlerMock) ValidateToken(appName, token string) (response.OAuth, error) {
	oa.ValTknClld = true
	return &facebook.Response{
		ResponseData: facebook.ResponseData{
			Valid: oa.ExpValid,
			UsrID: oa.AppUserID,
		},
	}, oa.ExpErr
}

type PhoneVerifierMock struct {
	ExpErr    error
	ExpStatus *authms.SMSVerificationStatus
}

func (pv *PhoneVerifierMock) SendSMSCode(toPhone string) (*authms.SMSVerificationStatus, error) {
	return pv.ExpStatus, pv.ExpErr
}
func (pv *PhoneVerifierMock) VerifySMSCode(r *authms.SMSVerificationCodeRequest) (*authms.SMSVerificationStatus, error) {
	return pv.ExpStatus, pv.ExpErr
}

type DBHelperMock struct {
	ExpErr                     error
	ExpUser                    *authms.User
	ExpHist                    []*authms.History
	SaveUserCalled             bool
	SaveHistoryCalled          bool
	GetUserCalled              bool
	SaveTokenCalled            bool
	UpdatePhoneCalled          bool
	UpdatePassCalled           bool
	UpdateAppUserIDCalled      bool
	T                          *testing.T
	IgnoreDefaultVerifiedCheck bool
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
func (d *DBHelperMock) GetByAppUserID(appName, appUserID string) (*authms.User, error) {
	d.GetUserCalled = true
	return d.ExpUser, d.ExpErr
}
func (d *DBHelperMock) SaveToken(*token.Token) error {
	d.SaveTokenCalled = true
	return d.ExpErr
}
func (d *DBHelperMock) GetHistory(userID int64, offset, count int, accessType ...string) ([]*authms.History, error) {
	return d.ExpHist, d.ExpErr
}
func (d *DBHelperMock) SaveHistory(*authms.History) error {
	d.SaveHistoryCalled = true
	return d.ExpErr
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

func (d *DBHelperMock)UpdateAppUserID(userID int64, new *authms.OAuth) error {
	d.UpdateAppUserIDCalled = true
	if new.Verified == false {
		d.T.Error("expected OAuth verified=true but got false")
	}
	return d.ExpErr
}

type RegisterTestCase struct {
	Desc      string
	User      *authms.User
	OAHandler *OAuthHandlerMock
	DBHelper  *DBHelperMock
	ExpErr    bool
	DevID     string
	Token     *token.Token
}

var conf = &ConfigMock{}
var tokenGen *token.Generator
var confFile = flag.String("conf", "/etc/authms/authms.conf.yml",
	"/path/to/config/file.conf.yml")

func init() {
	flag.Parse()
}

func TestNew(t *testing.T) {
	setUp(t)
	newAuth(t, &DBHelperMock{T: t}, &OAuthHandlerMock{}, &PhoneVerifierMock{})
}

func TestAuth_Register(t *testing.T) {
	appName := "test-app"
	cases := []RegisterTestCase{
		{
			Desc: "Valid args",
			User: &authms.User{ID: 123, UserName:"some-name", Password:"some-password"},
			OAHandler: &OAuthHandlerMock{},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: false,
			DevID: "test-dev",
		},
		{
			Desc: "Valid args, verified=true",
			User: &authms.User{
				ID: 123,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName: appName,
						AppUserID: "1",
						Verified:true,
						AppToken: "some-app-token",
					},
				},
				Phone: &authms.Value{Verified:true},
				Email: &authms.Value{Verified:true},
			},
			OAHandler: &OAuthHandlerMock{AppUserID:"1", ExpValid:true},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: false,
			DevID: "test-dev",
		},
		{
			Desc: "Nil user",
			User: nil, OAHandler: &OAuthHandlerMock{},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: "test-dev"},
		{
			Desc: "OAuth user",
			User: &authms.User{
				ID: 123,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName: appName,
						AppUserID: "test-oauth-app-uid",
						AppToken: "some-test-app-token",
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpValTknClld: true,
				ExpValid:true,
				AppUserID: "test-oauth-app-uid",
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: false,
			DevID: "test-dev",
		},
		{Desc: "Missing DevID", User: &authms.User{}, OAHandler: &OAuthHandlerMock{},
			DBHelper: &DBHelperMock{T: t}, ExpErr: true},
	}
	for _, c := range cases {
		func() {
			setUp(t)
			a := newAuth(t, c.DBHelper, c.OAHandler, &PhoneVerifierMock{})
			err := a.Register(c.User, c.DevID, "")
			runtime.Gosched()
			if c.OAHandler.ExpValTknClld && !c.OAHandler.ValTknClld {
				t.Errorf("%s - validate oath token not called", c.Desc)
			}
			if c.ExpErr {
				if err == nil {
					t.Errorf("%s - Expected an error but" +
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

func TestAuth_UpdatePhone(t *testing.T) {
	usrID := int64(123)
	devID := "test-dev"
	tkn, err := tokenGen.Generate(int(usrID), devID, token.ShortExpType)
	if err != nil {
		t.Fatalf("Error setting up: token.Generate(): %v", err)
	}
	cases := []RegisterTestCase{
		{
			Desc: "Valid phone",
			User: &authms.User{
				ID: usrID,
				Phone: &authms.Value{Value:"+254123456789"},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: false,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "Valid phone and verified=true (exp verified=false sent to dbhelper)",
			User: &authms.User{
				ID: usrID,
				Phone: &authms.Value{
					Value:"+254123456789",
					Verified:true},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: false,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "Invalid phone",
			User: &authms.User{
				ID: usrID,
				Phone: &authms.Value{},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "nil phone",
			User: &authms.User{
				ID: usrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "nil user",
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "no userID",
			User: &authms.User{
				Phone: &authms.Value{Value:"+254123456789"},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "Invalid token",
			User: &authms.User{
				ID: usrID + 5,
				Phone: &authms.Value{Value:"+254123456789"},
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
	}
	for _, c := range cases {
		func() {
			setUp(t)
			a := newAuth(t, c.DBHelper, c.OAHandler, &PhoneVerifierMock{})
			err := a.UpdatePhone(c.User, c.Token.Token(), c.DevID, "")
			runtime.Gosched()
			if c.ExpErr {
				if err == nil {
					t.Errorf("%s - Expected an error but" +
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
	tkn, err := tokenGen.Generate(int(usrID), devID, token.ShortExpType)
	if err != nil {
		t.Fatalf("Error setting up: token.Generate(): %v", err)
	}
	appName := "test-app"
	appUsrID := "test-app-userID"
	cases := []RegisterTestCase{
		{
			Desc: "Valid OAuth",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName: appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr: nil,
				ExpValid: true,
				AppUserID: appUsrID,
				ExpValTknClld: true,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: false,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "Mismatched OAuth names",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					"mismatched-name": {
						AppName: "mismatched-name",
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpValTknClld: false,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
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
				ExpErr: nil,
				ExpValid: true,
				AppUserID: appUsrID,
				ExpValTknClld: true,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "Missing user ID",
			User: &authms.User{
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName: appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr: nil,
				ExpValid: true,
				AppUserID: appUsrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "Nil OAuth",
			User: &authms.User{
				ID: usrID,
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr: nil,
				ExpValid: true,
				AppUserID: appUsrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "Valid OAuth and verified=true (exp dbhelper to be passed to false)",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName: appName,
						AppUserID: appUsrID,
						Verified: true,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr: nil,
				ExpValid: true,
				AppUserID: appUsrID,
				ExpValTknClld: true,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: false,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "Invalid token",
			User: &authms.User{
				ID: usrID + 5,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName: appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr: nil,
				ExpValid: true,
				AppUserID: appUsrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "OAuth report error",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName: appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr: errors.New(""),
				ExpValid: true,
				AppUserID: appUsrID,
				ExpValTknClld: true,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: devID,
			Token: tkn,
		},
		{
			Desc: "Missing device ID",
			User: &authms.User{
				ID: usrID,
				OAuths: map[string]*authms.OAuth{
					appName: {
						AppName: appName,
						AppUserID: appUsrID,
					},
				},
			},
			OAHandler: &OAuthHandlerMock{
				ExpErr: nil,
				ExpValid: true,
				AppUserID: appUsrID,
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: true,
			DevID: "",
			Token: tkn,
		},
	}
	for _, c := range cases {
		func() {
			setUp(t)
			a := newAuth(t, c.DBHelper, c.OAHandler, &PhoneVerifierMock{})
			err := a.UpdateOAuth(c.User, appName, c.Token.Token(), c.DevID, "")
			runtime.Gosched()
			if c.OAHandler.ExpValTknClld && !c.OAHandler.ValTknClld {
				t.Errorf("%s - validate token was not called", c.Desc)
			}
			if c.ExpErr {
				if err == nil {
					t.Errorf("%s - Expected an error but" +
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
		ExpErr bool
		Desc   string
		Req    *authms.SMSVerificationRequest
		PV     *PhoneVerifierMock
	}
	phone := "+254712345678"
	devID := "test-dev"
	userID := int64(1234)
	validToken, err := tokenGen.Generate(int(userID), devID, token.ShortExpType)
	if err != nil {
		t.Fatalf("error setting up (generatng token): %v", err)
	}
	tcs := []VerifyTestCase{
		{
			Desc: "Successful validation",
			ExpErr: false,
			Req: &authms.SMSVerificationRequest{
				DeviceID: devID,
				UserID: userID,
				Token: validToken.Token(),
				Phone: phone,
			},
			PV: &PhoneVerifierMock{
				ExpErr: nil,
				ExpStatus: &authms.SMSVerificationStatus{
					Token: "some-sms-token",
					Phone: phone,
					ExpiresAt: "2017-02-20T22:07:00+03:00",
					Verified: false,
				},
			},
		},
	}
	for _, tc := range tcs {
		func() {
			setUp(t)
			db := &DBHelperMock{T: t}
			a := newAuth(t, db, &OAuthHandlerMock{}, tc.PV)
			r, err := a.VerifyPhone(tc.Req, "")
			runtime.Gosched()
			if tc.ExpErr {
				if err == nil {
					t.Errorf("%s - expected error but " +
						"got nil", tc.Desc)
				}
				return
			} else if err != nil {
				t.Errorf("%s - auth.VerifyPhone(): %v", tc.Desc, err)
				return
			}
			if !db.SaveHistoryCalled {
				t.Errorf("%s - save history was not called", tc.Desc)
			}
			if r != tc.PV.ExpStatus {
				t.Errorf("%s\ngot status %+v\nexpected %+v",
					tc.Desc, r, tc.PV.ExpStatus)
				return
			}
		}()
	}
}

func TestAuth_VerifyPhoneCode(t *testing.T) {
	type VerifyTestCase struct {
		ExpErr bool
		Desc   string
		Req    *authms.SMSVerificationCodeRequest
		PV     *PhoneVerifierMock
		DB     *DBHelperMock
	}
	phone := "+254712345678"
	devID := "test-dev"
	userID := int64(1234)
	validToken, err := tokenGen.Generate(int(userID), devID, token.ShortExpType)
	if err != nil {
		t.Fatalf("error setting up (generatng token): %v", err)
	}
	tcs := []VerifyTestCase{
		{
			Desc: "Successful validation",
			ExpErr: false,
			Req: &authms.SMSVerificationCodeRequest{
				DeviceID: devID,
				UserID: userID,
				SmsToken: "some-sms-token",
				Token: validToken.Token(),
			},
			PV: &PhoneVerifierMock{
				ExpErr: nil,
				ExpStatus: &authms.SMSVerificationStatus{
					Phone: phone,
					Verified: true,
				},
			},
		},
	}
	for _, tc := range tcs {
		func() {
			setUp(t)
			db := &DBHelperMock{T: t, IgnoreDefaultVerifiedCheck: true}
			a := newAuth(t, db, &OAuthHandlerMock{}, tc.PV)
			r, err := a.VerifyPhoneCode(tc.Req, "")
			runtime.Gosched()
			if tc.ExpErr {
				if err == nil {
					t.Errorf("%s - expected error but " +
						"got nil", tc.Desc)
				}
				return
			} else if err != nil {
				t.Errorf("%s - auth.VerifyPhone(): %v", tc.Desc, err)
				return
			}
			if !db.SaveHistoryCalled {
				t.Errorf("%s - save history was not called", tc.Desc)
			}
			if !db.UpdatePhoneCalled {
				t.Errorf("%s - update phone was not called", tc.Desc)
			}
			if r != tc.PV.ExpStatus {
				t.Errorf("%s\ngot status %+v\nexpected %+v",
					tc.Desc, r, tc.PV.ExpStatus)
				return
			}
		}()
	}
}

func TestAuth_UpdatePassword(t *testing.T) {
	usrID := int64(123)
	devID := "test-dev"
	cases := []RegisterTestCase{
		{
			Desc: "Valid password",
			User: &authms.User{
				ID: usrID,
				Password: "some-new-password",
			},
			DBHelper: &DBHelperMock{T: t},
			ExpErr: false,
			DevID: devID,
		},
	}
	for _, c := range cases {
		func() {
			setUp(t)
			a := newAuth(t, c.DBHelper, c.OAHandler, &PhoneVerifierMock{})
			err := a.UpdatePassword(c.User.ID, "some-old-pass",
				c.User.Password, c.DevID, "")
			runtime.Gosched()
			if c.ExpErr {
				if err == nil {
					t.Errorf("%s - Expected an error but" +
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
		{Desc:"Valid Creds", ExpErr:false,
			DBHelper:&DBHelperMock{ExpUser:&authms.User{ID:123}},
			OAHandler: &OAuthHandlerMock{ExpValTknClld:true,
				AppUserID:"test-app-user-id", ExpValid: true},
			OAuth: &authms.OAuth{AppUserID:"test-app-user-id"}, DevID:"Tes-devID"},
		{Desc:"Invalid OAuth Creds", ExpErr:true,
			DBHelper:&DBHelperMock{T: t},
			OAHandler: &OAuthHandlerMock{ExpValTknClld:true, ExpErr: errors.New("")},
			OAuth: &authms.OAuth{}, DevID:"Tes-devID"},
	}
	for _, c := range cases {
		func() {
			a := newAuth(t, c.DBHelper, c.OAHandler, &PhoneVerifierMock{})
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
			if !c.DBHelper.SaveTokenCalled {
				t.Errorf("%s - save token not called", c.Desc)
			}
		}()
	}
}

func newAuth(t *testing.T, db *DBHelperMock, oa *OAuthHandlerMock, pv *PhoneVerifierMock) *auth.Auth {
	var err error
	tokenGen, err = token.NewGenerator(conf.Token)
	if err != nil {
		t.Fatalf("token.NewGenerator(): %v", err)
	}
	lg := log4go.NewDefaultLogger(log4go.FINEST)
	a, err := auth.New(tokenGen, lg, db, oa, pv)
	if err != nil {
		t.Fatalf("auth.New(): %v", err)
	}
	return a
}

func setUp(t *testing.T) {
	if err := config.ReadYamlConfig(*confFile, conf); err != nil {
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