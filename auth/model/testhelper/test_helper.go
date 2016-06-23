package testhelper

import (
	"database/sql"
	"fmt"
	"testing"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	_ "github.com/lib/pq"
)

const (
	DBName = "test_authms"
)

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
