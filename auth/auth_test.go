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
	TknKeyFile string `yaml:"tokenkeyfile"`
}

func (c TokenConfigMock) TokenKeyFile() string {
	return c.TknKeyFile
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

type DBHelperMock struct {
	ExpErr            error
	ExpUser           *authms.User
	ExpHist           []*authms.History
	SaveUserCalled    bool
	SaveHistoryCalled bool
	GetUserCalled     bool
	SaveTokenCalled   bool
}

func (d *DBHelperMock) SaveUser(*authms.User) error {
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
func (d *DBHelperMock) GetHistory(userID int64, offset, count int, accessType string) ([]*authms.History, error) {
	return d.ExpHist, d.ExpErr
}
func (d *DBHelperMock) SaveHistory(*authms.History) error {
	d.SaveHistoryCalled = true
	return d.ExpErr
}

var conf = &ConfigMock{}
var confFile = flag.String("conf", "/etc/authms/authms.conf.yml",
	"/path/to/config/file.conf.yml")

func init() {
	flag.Parse()
}

func TestNew(t *testing.T) {
	setUp(t)
	newAuth(t, &DBHelperMock{}, &OAuthHandlerMock{})
}

func TestAuth_Register(t *testing.T) {
	type RegisterTestCase struct {
		Desc      string
		User      *authms.User
		OAHandler *OAuthHandlerMock
		DBHelper  *DBHelperMock
		ExpErr    bool
		DevID     string
	}
	cases := []RegisterTestCase{
		{
			Desc: "Valid args",
			User: &authms.User{ID: 123},
			OAHandler: &OAuthHandlerMock{},
			DBHelper: &DBHelperMock{},
			ExpErr: false,
			DevID: "test-dev",
		},
		{
			Desc: "Nil user",
			User: nil, OAHandler: &OAuthHandlerMock{},
			DBHelper: &DBHelperMock{},
			ExpErr: true,
			DevID: "test-dev"},
		{
			Desc: "OAuth user",
			User: &authms.User{
				ID: 123,
				OAuth:&authms.OAuth{AppUserID: "test-oauth-app-uid"},
			},
			OAHandler: &OAuthHandlerMock{
				ExpValTknClld: true,
				ExpValid:true,
				AppUserID: "test-oauth-app-uid",
			},
			DBHelper: &DBHelperMock{},
			ExpErr: false,
			DevID: "test-dev",
		},
		{Desc: "Missing DevID", User: &authms.User{}, OAHandler: &OAuthHandlerMock{},
			DBHelper: &DBHelperMock{}, ExpErr: true},
	}
	for _, c := range cases {
		func() {
			setUp(t)
			a := newAuth(t, c.DBHelper, c.OAHandler)
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
			DBHelper:&DBHelperMock{},
			OAHandler: &OAuthHandlerMock{ExpValTknClld:true, ExpErr: errors.New("")},
			OAuth: &authms.OAuth{}, DevID:"Tes-devID"},
	}
	for _, c := range cases {
		func() {
			a := newAuth(t, c.DBHelper, c.OAHandler)
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
			if !c.DBHelper.GetUserCalled {
				t.Errorf("%s - get user not called", c.Desc)
			}
			if !c.DBHelper.SaveTokenCalled {
				t.Errorf("%s - save token not called", c.Desc)
			}
		}()
	}
}

func newAuth(t *testing.T, db *DBHelperMock, oa *OAuthHandlerMock) *auth.Auth {
	tg, err := token.NewGenerator(conf.Token)
	if err != nil {
		t.Fatalf("token.NewGenerator(): %v", err)
	}
	lg := log4go.NewDefaultLogger(log4go.FINEST)
	a, err := auth.New(tg, lg, db, oa)
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