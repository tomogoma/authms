package testhelper

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/tomogoma/authms/auth/model/helper"
)

const (
	DBName = "test_authms"
)

type App struct {
	AppUID     string
	AppName    string
	AppToken   string
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

func (a *App) Token() string {
	return a.AppToken
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

func HashF(p string) ([]byte, error) {
	return []byte{0, 1, 2, 3, 4, 5}, nil
}

func ValHashFunc(p string, passHB []byte) bool {
	return true
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
