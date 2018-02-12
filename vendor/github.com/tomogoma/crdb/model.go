package crdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb"
	_ "github.com/lib/pq"
)

type DSNFormatter interface {
	FormatDSN() string
}

const (
	driverName = "postgres"
)

// DBConn establishes a connection to the database.
func DBConn(dsn string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	return CloseDBOnError(db, db.Ping())
}

// InstantiateDB creates the database and tables based on dbName and tableDescs.
func InstantiateDB(db *sql.DB, dbName string, tableDescs ...string) error {
	return instantiateDB(db, dbName, true, tableDescs...)
}

// TryConnect attempts to connect to db using dsn if db is nil.
func TryConnect(dsn string, db *sql.DB) (*sql.DB, error) {
	if db != nil {
		return db, nil
	}
	return DBConn(dsn)
}

// CloseDBOnError closes connection to the given database if err != nil.
// It then returns the error.
func CloseDBOnError(db *sql.DB, err error) (*sql.DB, error) {
	if err == nil {
		return db, nil
	}
	if db == nil {
		return nil, err
	}
	clErr := db.Close()
	if clErr != nil {
		return nil, fmt.Errorf("%v ...and close DB: %v",
			err, clErr)
	}
	return nil, err
}

func instantiateDB(db *sql.DB, dbName string, createDB bool, tableDescs ...string) error {
	retryWithoutCreateDB := false
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := crdb.ExecuteTx(ctx, db, nil, func(tx *sql.Tx) error {
		if createDB {
			// This block is necessary to accommodate non-root user
			// as CockroachDB only allows root user to CREATE DATABASE.
			_, err := tx.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
			if err != nil {
				retryWithoutCreateDB = true
				return err
			}
		}
		for _, createTblStmt := range tableDescs {
			_, err := tx.Exec(createTblStmt)
			if err != nil {
				return fmt.Errorf("error creating table: %v", err)
			}
		}
		return nil
	})
	if err != nil && retryWithoutCreateDB {
		return instantiateDB(db, dbName, false, tableDescs...)
	}
	return err
}
