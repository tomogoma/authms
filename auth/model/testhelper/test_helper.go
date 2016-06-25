package testhelper

import (
	"database/sql"
	"fmt"
	"testing"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/history"
	"bitbucket.org/tomogoma/auth-ms/auth/model/token"
	"bitbucket.org/tomogoma/auth-ms/auth/model/user"
	_ "github.com/lib/pq"
)

const (
	DBName = "test_authms"
)

type User struct {
	UName    string
	Password string

	FName       string
	MName       string
	LName       string
	HashF       user.HashFunc
	ValHashFunc user.ValidatePassFunc
}

func (u *User) ID() int                            { return 1 }
func (u *User) UserName() string                   { return u.UName }
func (u *User) FirstName() string                  { return u.FName }
func (u *User) MiddleName() string                 { return u.MName }
func (u *User) LastName() string                   { return u.LName }
func (u *User) PreviousLogins() []*history.History { return make([]*history.History, 0) }
func (u *User) Token() token.Token                 { t, _ := token.New(1, "test", token.ShortExpType); return t }

func HashF(p string) ([]byte, error) {
	return []byte{0, 1, 2, 3, 4, 5}, nil
}

func ValHashFunc(p string, passHB []byte) bool {
	return true
}

func (u User) ExplodeParams() (string, string, string, string, string, user.HashFunc) {
	return u.UName, u.FName, u.MName, u.LName, u.Password, u.HashF
}

var DSN = helper.DSN{
	UName:       "root",
	Host:        "z500:26257",
	SslCert:     "/etc/cockroachdb/certs/node.cert",
	SslKey:      "/etc/cockroachdb/certs/node.key",
	SslRootCert: "/etc/cockroachdb/certs/ca.cert",
}

func SetUp(m helper.Model, db *sql.DB, t *testing.T) {

	if db == nil {
		t.Fatalf("Found nil db while setting up")
	}

	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", DBName))
	if err != nil {
		db.Close()
		t.Fatalf("Error creating test database: %s", err)
	}

	_, err = db.Exec(fmt.Sprintf("SET DATABASE = %s", DBName))
	if err != nil {
		TearDown(db, t)
		t.Fatalf("Error setting default db to test database: %s", err)
	}

	if m == nil {
		return
	}

	_, err = db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", m.TableName(), m.TableDesc()))
	if err != nil {
		TearDown(db, t)
		t.Fatalf("creating test table: %s", err)
	}
}

func TearDown(db *sql.DB, t *testing.T) {

	if db == nil {
		t.Logf("Found nil db while tearing down")
		return
	}

	defer db.Close()
	_, err := db.Exec(fmt.Sprintf("DROP DATABASE %s", DBName))
	if err != nil {
		t.Errorf("dropping test db: %s", err)
	}
}

func InstantiateDB(t *testing.T) *sql.DB {

	db, err := helper.SQLDB(DSN)
	if err != nil {
		t.Fatalf("helper.SQLDB(): %s", err)
	}

	if db == nil {
		t.Fatalf("Expected db but got nil")
	}

	return db
}
