package testhelper

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/tomogoma/authms/auth/model/helper"
	"github.com/tomogoma/authms/auth/model/history"
	"github.com/tomogoma/authms/auth/model/token"
	"github.com/tomogoma/authms/auth/model/user"
	"github.com/tomogoma/authms/auth/password"
)

const (
	DBName = "test_authms"
)

type App struct {
	AppUID     string
	AppName    string
	IsVerified bool
}

func (a *App) UserID() string {
	return a.AppUID
}

func (a *App) Name() string {
	if a == nil {
		return ""
	}
	return a.AppName
}

func (a *App) Validated() bool {
	return a.IsVerified
}

type Value struct {
	Val        string
	IsVerified bool
}

func (v *Value) Value() string {
	return v.Val
}

func (v *Value) Validated() bool {
	return v.IsVerified
}

type User struct {
	UName     string
	EmailAddr *Value
	PhoneNo   *Value
	AppDet    *App

	Password    string
	HashF       user.HashFunc
	ValHashFunc user.ValidatePassFunc
	TokenGen    *token.Generator
}

func (u *User) ID() int {
	return 1
}
func (u *User) UserName() string {
	return u.UName
}
func (u *User) Email() user.Valuer {
	return u.EmailAddr
}
func (u *User) EmailAddress() string {
	if u.EmailAddr == nil {
		return ""
	}
	return u.EmailAddr.Val
}
func (u *User) Phone() user.Valuer {
	return u.PhoneNo
}
func (u *User) PhoneNumber() string {
	if u.PhoneNo == nil {
		return ""
	}
	return u.PhoneNo.Val
}
func (u *User) App() user.App {
	return u.AppDet
}
func (u *User) PreviousLogins() []*history.History {
	return make([]*history.History, 0)
}
func (u *User) Token() token.Token {
	t, _ := u.TokenGen.Generate(1, "test", token.ShortExpType)
	return t
}

func HashF(p string) ([]byte, error) {
	return []byte{0, 1, 2, 3, 4, 5}, nil
}

func ValHashFunc(p string, passHB []byte) bool {
	return true
}

func (u User) ExplodeParams() (string, string, string, string, user.App, *password.Generator, user.HashFunc) {
	g, _ := password.NewGenerator(password.AllChars)
	return u.UName, u.PhoneNo.Val, u.EmailAddr.Val, u.Password, u.AppDet, g, u.HashF
}

var DSN = helper.DSN{
	UName:       "root",
	Host:        "localhost:26257",
	SslCert:     "/etc/cockroachdb/certs/node.cert",
	SslKey:      "/etc/cockroachdb/certs/node.key",
	SslRootCert: "/etc/cockroachdb/certs/ca.cert",
	DB:          "test_authms",
}

func TearDown(db *sql.DB, t *testing.T) {
	if db == nil {
		t.Log("Found nil db while tearing down")
		return
	}
	tables := []string{"tokens", "history", "userNames", "emails",
		"phones", "appUserIDs", "users"}
	for _, table := range tables {
		if _, err := db.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			t.Errorf("dropping table %s: %s", table, err)
		}
	}
	defer db.Close()
	_, err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", DBName))
	if err != nil {
		t.Errorf("dropping test db: %s", err)
	}
}

func SQLDB(t *testing.T) *sql.DB {
	db, err := helper.SQLDB(DSN)
	if err != nil {
		t.Fatalf("helper.SQLDB(): %s", err)
	}
	if db == nil {
		t.Fatal("Expected db but got nil")
	}
	return db
}

func InsertDummyUser(db *sql.DB, userID int, t *testing.T) {
	insertQ := `
	INSERT INTO users (id, password, createDate)
	 	VALUES ($1, $2, CURRENT_TIMESTAMP())
	 `
	_, err := db.Exec(insertQ, userID, []byte("SOME-PASSWORD-HASH"))
	if err != nil {
		TearDown(db, t)
		t.Fatalf("Error setting up dummy user: %s", err)
	}
}
