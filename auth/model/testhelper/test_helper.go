package testhelper

import (
	"database/sql"
	"fmt"
	"testing"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	_ "github.com/lib/pq"
)

const (
	dbname       = "test_authms"
	dbDriverName = "postgres"
)

func SetUp(m helper.Model, db *sql.DB, t *testing.T) {

	if db == nil {
		t.Fatalf("Found nil db while setting up")
	}

	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	if err != nil {
		TearDown(db, t)
		t.Fatalf("creating test db: %s", err)
	}

	_, err = db.Exec(fmt.Sprintf("SET DATABASE = %s", dbname))
	if err != nil {
		TearDown(db, t)
		t.Fatalf("selecting test db: %s", err)
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
	_, err := db.Exec(fmt.Sprintf("DROP DATABASE %s", dbname))
	if err != nil {
		t.Errorf("dropping test db: %s", err)
	}
}

func InstantiateDB(t *testing.T) *sql.DB {

	dsn := "postgresql://root@z500:26257?sslcert=%2Fetc%2Fcockroachdb%2Fcerts%2Fnode.cert&sslkey=%2Fetc%2Fcockroachdb%2Fcerts%2Fnode.key&sslmode=verify-full&sslrootcert=%2Fetc%2Fcockroachdb%2Fcerts%2Fca.cert"

	db, err := sql.Open(dbDriverName, dsn)
	if err != nil {
		t.Fatalf("sql.Open(): %s", err)
	}

	if err = db.Ping(); err != nil {
		t.Fatalf("db.Ping(): %s", err)
	}

	return db
}
