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

func SetUp(m helper.Model, db *sql.DB, t *testing.T) {

	if db == nil {
		t.Fatalf("Found nil db while setting up")
	}

	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", DBName))
	if err != nil {
		TearDown(db, t)
		t.Fatalf("creating test db: %s", err)
	}

	_, err = db.Exec(fmt.Sprintf("SET DATABASE = %s", DBName))
	if err != nil {
		TearDown(db, t)
		t.Fatalf("selecting test db: %s", err)
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

	dsn := helper.DSN{
		UName:       "root",
		Host:        "z500:26257",
		SslCert:     "/etc/cockroachdb/certs/node.cert",
		SslKey:      "/etc/cockroachdb/certs/node.key",
		SslRootCert: "/etc/cockroachdb/certs/ca.cert",
	}

	db, err := helper.SQLDB(dsn)
	if err != nil {
		t.Fatalf("helper.SQLDB(): %s", err)
	}

	if db == nil {
		t.Fatalf("Expected db but got nil")
	}

	return db

	//db, err := sql.Open(dbDriverName, DSN.FormatDSN())
	//if err != nil {
	//	t.Fatalf("sql.Open(): %s", err)
	//}
	//
	//if err = db.Ping(); err != nil {
	//	t.Fatalf("db.Ping(): %s", err)
	//}
	//
	//return db
}
