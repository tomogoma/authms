package auth_test

import (
	"testing"

	"database/sql"

	"time"

	"github.com/limetext/log4go"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/model/history"
	"github.com/tomogoma/authms/auth/model/testhelper"
	"github.com/tomogoma/authms/auth/model/token"
	"github.com/tomogoma/authms/auth/model/user"
	"github.com/tomogoma/authms/auth/password"
)

type History struct {
	accessSuccessStatus bool
	accessType          int
	isSaveCalled        bool

	testPoint    int
	currentPoint int

	desc string
	t    *testing.T
}

func (h *History) TableName() string {
	return "history"
}

func (h *History) TableDesc() string {
	return "id INT PRIMARY KEY"
}

func (h *History) Save(ld history.History) (int, error) {

	h.currentPoint++
	if h.testPoint != h.currentPoint {
		return 1, nil
	}

	h.isSaveCalled = true
	if h.accessSuccessStatus != ld.Successful() {
		h.t.Errorf("Test %s: expected successfull (%v) but got (%v)",
			h.desc, h.accessSuccessStatus, ld.Successful())
	}

	if h.accessType != ld.AccessMethod() {
		h.t.Errorf("Test %s: expected save type (%v) but got (%v)",
			h.desc, h.accessType, ld.AccessMethod())
	}
	return 1, nil
}

func (h *History) Get(userID, offset, count int, acMs ...int) ([]*history.History, error) {
	hs, err := history.New(userID, h.accessType, h.accessSuccessStatus, time.Now(), "127.0.0.1", "test", "auth")
	if err != nil {
		h.t.Errorf("Test %s: history.New(): %s", err)
	}
	return []*history.History{hs}, nil
}

var db *sql.DB
var histM = &History{
	accessType:          history.RegistrationAccess,
	accessSuccessStatus: true,
}

var tokenGen *token.Generator

func TestNew(t *testing.T) {
	newAuth(histM, t)
	defer testhelper.TearDown(db, t)
}

func TestAuth_RegisterUser(t *testing.T) {

	type testcase struct {
		expErr error
		desc   string
		user   *testhelper.User
		hist   *History
	}

	cases := []testcase{
		{
			expErr: nil,
			desc:   "register successfull",
			user: &testhelper.User{
				UName:    "uname",
				Password: "pass",
				TokenGen: tokenGen,
			},
			hist: &History{
				desc:                "register successfull",
				testPoint:           1,
				accessType:          history.RegistrationAccess,
				accessSuccessStatus: true,
				t:                   t,
			},
		},
		{
			expErr: user.ErrorEmptyIdentifier,
			desc:   "register unsuccessful",
			user: &testhelper.User{
				Password: "pass",
				TokenGen: tokenGen,
			},
			hist: &History{
				desc:                "register unsuccessful",
				testPoint:           1,
				accessType:          history.RegistrationAccess,
				accessSuccessStatus: false,
				t:                   t,
			},
		},
	}

	for _, c := range cases {
		func() {
			a := newAuth(c.hist, t)
			defer testhelper.TearDown(db, t)

			_, err := a.RegisterUser(c.user, c.user.Password,
				"127.0.0.1", "authms", "test")
			if err != c.expErr {
				t.Errorf("Test %s: auth.RegisterUser(): "+
					"expected error %v got %v",
					c.desc, c.expErr, err)
				return
			}

			if !c.hist.isSaveCalled {
				t.Errorf("Test %s: save history was never"+
					" called", c.desc)
				return
			}

			if c.expErr != nil {
				return
			}

			_, err = a.LoginUserName(c.user.UserName(),
				c.user.Password, "devid", "127.0.0.1",
				"tester", "auth")
			if err != nil {
				t.Errorf("Test %s: auth.Login(): %s",
					c.desc, err)
			}
		}()
	}
}

func TestAuth_RegisterUser_limitsFailure(t *testing.T) {
	//
	//a := newAuth(histM, t)
	//usr := &testhelper.User{}
	//a.RegisterUser(usr, "pass", "127.0.0.1", "test", "authms")
	//a.RegisterUser(usr, "pass", "127.0.0.1", "test", "authms")
	//a.RegisterUser(usr, "pass", "127.0.0.1", "test", "authms")
	//actUsr := testhelper.User{UName: "uname", Password: "pass"}
	//svdUsr, err := a.RegisterUser(actUsr, "pass", "127.0.0.1", "test", "authms")
	//if err == nil || err !=
	// TODO p>m login/token attempts from an ip address over duration t
	// 1. Check db if failed attempts of each type login, reg, token were made in
	//    the past configured max time
	// 2. check, for each type, if the failed count exceeds the configured max
	// 3. if it does, return specific error
	t.Fatal("untested:\nCheck:\n3 failed attempts for a user from an IP in an hour blocked\n6 attempts from a user in an hour blocked")
}

func TestAuth_Login(t *testing.T) {

	type testcase struct {
		expErr error
		desc   string
		user   *testhelper.User
		hist   *History
	}

	regdUsr := &testhelper.User{
		UName:    "uname",
		Password: "pass",
		TokenGen: tokenGen,
	}

	cases := []testcase{
		{
			expErr: nil,
			desc:   "login successfull",
			user:   regdUsr,
			hist: &History{
				desc:                "login successfull",
				testPoint:           2,
				accessType:          history.LoginAccess,
				accessSuccessStatus: true,
				t:                   t,
			},
		},
		{
			expErr: user.ErrorPasswordMismatch,
			desc:   "login unsuccessful",
			user:   &testhelper.User{Password: "pass"},
			hist: &History{
				desc:                "login unsuccessfull",
				testPoint:           2,
				accessType:          history.LoginAccess,
				accessSuccessStatus: false,
				t:                   t,
			},
		},
	}

	for _, c := range cases {
		func() {
			a := newAuth(c.hist, t)
			defer testhelper.TearDown(db, t)

			_, err := a.RegisterUser(regdUsr, regdUsr.Password, "127.0.0.1", "authms", "test")
			if err != nil {
				t.Fatalf("Test %s: auth.RegisterUser(): %s", c.desc, err)
				return
			}

			_, err = a.LoginUserName(c.user.UName, c.user.Password,
				"devid", "127.0.0.1", "tester", "auth")
			if err != c.expErr {
				t.Errorf("Test %s: auth.Login(): expected error %v but got %v",
					c.desc, c.expErr, err)
			}

			if !c.hist.isSaveCalled {
				t.Errorf("Test %s: save history was never called", c.desc)
				return
			}
		}()
	}
}

func TestAuth_Login2(t *testing.T) {
	// TODO p>m login/token attempts from an ip address over duration t
	t.Fatal("untested:\nCheck:\n3 failed attempts for a user from an IP in an hour blocked\n6 attempts from a user in an hour blocked")
}

func TestAuth_AuthenticateToken(t *testing.T) {

	type testcase struct {
		expErr  error
		desc    string
		useRegd bool
		hist    *History
	}

	regdUsr := &testhelper.User{
		UName:    "uname",
		Password: "pass",
		TokenGen: tokenGen,
	}

	cases := []testcase{
		{
			expErr:  nil,
			desc:    "token matches",
			useRegd: true,
			hist: &History{
				desc:                "token matches",
				testPoint:           3,
				accessType:          history.TokenValidationAccess,
				accessSuccessStatus: true,
				t:                   t,
			},
		},
		{
			expErr:  token.ErrorInvalidToken,
			desc:    "token mismatch",
			useRegd: false,
			hist: &History{
				desc:                "token mismatch",
				testPoint:           3,
				accessType:          history.TokenValidationAccess,
				accessSuccessStatus: false,
				t:                   t,
			},
		},
	}

	for _, c := range cases {
		func() {
			a := newAuth(c.hist, t)
			defer testhelper.TearDown(db, t)

			_, err := a.RegisterUser(regdUsr, regdUsr.Password, "127.0.0.1", "authms", "test")
			if err != nil {
				t.Fatalf("Test %s: auth.RegisterUser(): %s", c.desc, err)
				return
			}
			u, err := a.LoginUserName(regdUsr.UName, regdUsr.Password, "devid", "127.0.0.1", "tester", "auth")
			if err != nil {
				t.Fatalf("Test %s: auth.loginUserName(): %s", c.desc, err)
				return
			}
			tkStr := u.Token("")
			if !c.useRegd {
				tkn, err := tokenGen.Generate(regdUsr.ID(), "devid", token.ShortExpType)
				if err != nil {
					t.Errorf("Test %s: unable to generate dummy token: %s", c.desc, err)
				}
				tkStr = tkn.Token()
			}
			usr, err := a.AuthenticateToken(tkStr, "127.0.0.1", "tester", "auth")
			if err != c.expErr {
				t.Errorf("Test %s: auth.AuthenticateToken(): expected error %v but got %v", c.desc, c.expErr, err)
			}

			if !c.hist.isSaveCalled {
				t.Errorf("Test %s: save history was never called", c.desc)
			}

			if c.expErr != nil {
				return
			}

			if usr == nil {
				t.Errorf("Test %s: expected a user, got nil", c.desc)
				return
			}

			if usr.Token("") == "" {
				t.Errorf("Test %s: expected a token, got empty", c.desc)
			}

			if len(usr.PreviousLogins()) != 1 {
				t.Errorf("Test %s: expected 1 previous login, got %d", c.desc, len(usr.PreviousLogins()))
			}
		}()
	}
}

func TestAuth_AuthenticateToken2(t *testing.T) {
	// TODO p>m login/token attempts from an ip address over duration t
	t.Fatal("untested:\nCheck:\n3 failed attempts for a user from an IP in an hour blocked\n6 attempts from a user in an hour blocked")
}

func TestAPIKeysEnforced(t *testing.T) {
	t.Fatal("untested:\nAccess auth services only if client (microservice) / API key combo is recogonized")
}

func newAuth(h *History, t *testing.T) *auth.Auth {

	db = testhelper.SQLDB(t)

	quitCh := make(chan error)
	conf := auth.Config{
		BlacklistWindow:    30 * time.Minute,
		BlackListFailCount: 3,
	}

	var err error
	tokenGen, err = token.NewGenerator(token.Config{
		TokenKeyFile: "../ssh-keys/sha256.key",
	})
	if err != nil {
		t.Fatalf("New Token Generator: %s", err)
	}
	passGen, err := password.NewGenerator(password.AllChars)
	if err != nil {
		t.Fatalf("New Token Generator: %s", err)
	}

	lg := log4go.NewDefaultLogger(log4go.FINEST)
	a, err := auth.New(db, h, tokenGen, passGen, conf, lg, quitCh)
	if err != nil {
		t.Fatalf("auth.New(): %s", err)
	}

	if a == nil {
		t.Fatal("auth was nil, expected a value")
	}

	return a
}
