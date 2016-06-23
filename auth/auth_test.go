package auth_test

import (
	"testing"

	"database/sql"

	"bitbucket.org/tomogoma/auth-ms/auth"
	"bitbucket.org/tomogoma/auth-ms/auth/model/testhelper"
)

var db *sql.DB

func TestNew(t *testing.T) {
	newAuth(t)
	defer testhelper.TearDown(db, t)
}

func TestAuth_RegisterUser(t *testing.T) {
	//
	//a := newAuth(t)
	//usr, err := user.New("uname", "fname", "", "", "pass", user.Hash)
	//a.RegisterUser(usr, "pass")
	t.Fatalf("untested, check:\nsaved to db (success and fail)\n")
}

func TestAuth_RegisterUser2(t *testing.T) {
	// TODO p>m login/token attempts from an ip address over duration t
	t.Fatalf("untested:, check:\n3 failed attempts for a user from an IP in an hour blocked\n6 attempts from a user in an hour blocked")
}

func TestAuth_Login(t *testing.T) {
	t.Fatalf("untested, check:\nsaved to db (success and fail)\n")
}

func TestAuth_Login2(t *testing.T) {
	// TODO p>m login/token attempts from an ip address over duration t
	t.Fatalf("untested:, check:\n3 failed attempts for a user from an IP in an hour blocked\n6 attempts from a user in an hour blocked")
}

func TestAuth_AuthenticateToken(t *testing.T) {
	t.Fatalf("untested, check:\nsaved to db (success and fail)\n")
}

func TestAuth_AuthenticateToken2(t *testing.T) {
	// TODO p>m login/token attempts from an ip address over duration t
	t.Fatalf("untested:, check:\n3 failed attempts for a user from an IP in an hour blocked\n6 attempts from a user in an hour blocked")
}

func TestAPIKeysEnforced(t *testing.T) {
	t.Fatalf("Access auth services only if client (microservice) / API key combo is recogonized")
}

func newAuth(t *testing.T) *auth.Auth {

	db = testhelper.InstantiateDB(t)
	testhelper.SetUp(nil, db, t)

	quitCh := make(chan error)
	dsn := testhelper.DSN
	dsn.DB = testhelper.DBName
	a, err := auth.New(dsn, quitCh)
	if err != nil {
		t.Fatalf("auth.New(): %s", err)
	}

	if a == nil {
		t.Fatalf("auth was nil, expected a value")
	}

	return a
}
