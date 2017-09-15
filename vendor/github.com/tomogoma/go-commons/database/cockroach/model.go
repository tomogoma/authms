package cockroach

import (
	"database/sql"
	"errors"
	"fmt"

	"context"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb"
	_ "github.com/lib/pq"
)

type DSNFormatter interface {
	FormatDSN() string
	DBName() string
}

const (
	driverName = "postgres"
)

var ErrorNilDSNFormatter = errors.New("DSNFormatter cannot be nil")

// DBConn establishes a connection to the database.
func DBConn(dsnF DSNFormatter) (*sql.DB, error) {
	if dsnF == nil {
		return nil, ErrorNilDSNFormatter
	}
	db, err := sql.Open(driverName, dsnF.FormatDSN())
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if CloseDBOnError(db, err); err != nil {
		return nil, err

	}
	return db, nil
}

// InstantiateDB creates the database and tables based on dbName and tableDescs.
func InstantiateDB(db *sql.DB, dbName string, tableDescs ...string) error {
	return instantiateDB(db, dbName, true, tableDescs...)
}

// TryConnect attempts to connect to db using dsn (if not already connected).
func TryConnect(dsn string, db *sql.DB) (*sql.DB, error) {
	if db != nil {
		return db, nil
	}
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// CloseDBOnError closes connection to the given database if err != nil.
// It then returns the error.
func CloseDBOnError(db *sql.DB, err error) error {
	if db == nil {
		return fmt.Errorf("nil db while trying to close db on error '%v'", err)
	}
	if err != nil {
		clErr := db.Close()
		if clErr != nil {
			return fmt.Errorf("%s ...and while closing db: %s",
				err, clErr)
		}
		return err
	}
	return nil
}

func instantiateDB(db *sql.DB, dbName string, createDB bool, tableDescs ...string) error {
	retryWithoutCreateDB := false
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := crdb.ExecuteTx(ctx, db, nil, func(tx *sql.Tx) error {
		if createDB {
			// This block is necessary to accommodate non-root user to call
			// this func as CockroachDB only allows root user to
			// CREATE DATABASE.
			_, err := tx.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
			if err != nil {
				retryWithoutCreateDB = true
				return err
			}
		}
		_, err := tx.Exec("SET DATABASE = " + dbName)
		if err != nil {
			return fmt.Errorf("error setting database: %v", err)
		}
		for _, createTblStmt := range tableDescs {
			_, err = tx.Exec(createTblStmt)
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
