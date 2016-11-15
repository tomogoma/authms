package helper

import (
	"database/sql"
	"errors"

	"fmt"

	"github.com/cockroachdb/cockroach-go/crdb"
	_ "github.com/lib/pq"
)

const (
	driverName        = "postgres"
	NoResultsErrorStr = "sql: no rows in result set"
)

var ErrorNilDB = errors.New("db cannot be nil")
var ErrorNilDSNFormatter = errors.New("DSNFormatter cannot be nil")

const (
	users = `
CREATE TABLE IF NOT EXISTS users (
  id         SERIAL PRIMARY KEY NOT NULL,
  password   BYTES              NOT NULL,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP()
);
`
	usernames = `
CREATE TABLE IF NOT EXISTS userNames (
  userID     SERIAL             NOT NULL REFERENCES users (id),
  id         SERIAL		NOT NULL,
  userName   STRING UNIQUE      NOT NULL,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY	(userID, id)
);
`
	emails = `
CREATE TABLE IF NOT EXISTS emails (
  userID     SERIAL             NOT NULL REFERENCES users (id),
  id         SERIAL		NOT NULL,
  email      STRING UNIQUE      NOT NULL,
  validated  BOOL               NOT NULL DEFAULT FALSE,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY	(userID, id)
);
`
	phones = `
CREATE TABLE IF NOT EXISTS phones (
  userID     SERIAL             NOT NULL REFERENCES users (id),
  id         SERIAL		NOT NULL,
  phone      STRING UNIQUE      NOT NULL,
  validated  BOOL               NOT NULL DEFAULT FALSE,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY	(userID, id)
);
`
	appUserIDs = `
CREATE TABLE IF NOT EXISTS appUserIDs (
  userID     SERIAL             NOT NULL REFERENCES users (id),
  id         SERIAL		NOT NULL,
  appUserID  STRING             NOT NULL,
  appName    STRING             NOT NULL,
  validated  BOOL               NOT NULL DEFAULT FALSE,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY	(userID, id),
  UNIQUE     (appName, appUserID)
);
`
	// TODO enforce that devID, userID [, and appID??] as unique
	tokens = `
CREATE TABLE IF NOT EXISTS tokens (
  userID SERIAL             NOT NULL REFERENCES users (id),
  id     SERIAL		    NOT NULL,
  devID  STRING             NOT NULL,
  token  STRING UNIQUE      NOT NULL,
  issued TIMESTAMP          NOT NULL,
  expiry TIMESTAMP          NOT NULL,
  PRIMARY KEY	(userID, id)
);
`
	// TODO add error column
	history = `
CREATE TABLE IF NOT EXISTS history (
  userID       SERIAL             NOT NULL REFERENCES users (id),
  id           SERIAL		  NOT NULL,
  date         TIMESTAMP          NOT NULL,
  accessMethod INT                NOT NULL,
  successful   BOOL               NOT NULL,
  forServiceID STRING,
  ipAddress    STRING,
  referral     STRING,
  PRIMARY KEY	(userID, id),
  INDEX         history_UserDate_indx (userID, DATE )
);
`
)

func SQLDB(dsnF DSNFormatter) (*sql.DB, error) {
	if dsnF == nil {
		return nil, ErrorNilDSNFormatter
	}
	db, err := sql.Open(driverName, dsnF.FormatDSN())
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if closeDBOnError(db, err); err != nil {
		return nil, err

	}
	err = createTables(db, dsnF.DBName())
	if closeDBOnError(db, err); err != nil {
		return nil, err

	}
	return db, nil
}

func createTables(db *sql.DB, dbName string) error {
	return crdb.ExecuteTx(db, func(tx *sql.Tx) error {
		_, err := tx.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
		if err != nil {
			return err
		}
		_, err = tx.Exec("SET DATABASE = " + dbName)
		if err != nil {
			return err
		}
		createStmts := []string{users, usernames, emails, phones, appUserIDs, tokens, history}
		for _, createTblStmt := range createStmts {
			_, err = tx.Exec(createTblStmt)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func closeDBOnError(db *sql.DB, err error) error {
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
