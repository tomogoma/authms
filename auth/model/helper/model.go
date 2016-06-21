package helper

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	driverName        = "postgres"
	NoResultsErrorStr = "sql: no rows in result set"
)

var ErrorNilDB = errors.New("db cannot be nil")
var ErrorNilDSNFormatter = errors.New("DSNFormatter cannot be nil")

type Model interface {
	TableName() string
	TableDesc() string
}

func New(dsnF DSNFormatter) (*sql.DB, error) {

	if dsnF == nil {
		return nil, ErrorNilDSNFormatter
	}

	db, err := sql.Open(driverName, dsnF.FormatDSN())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTables(db *sql.DB, models ...Model) error {

	modelsM := make(map[string]string)
	for _, model := range models {

		tName := model.TableName()
		if _, ok := modelsM[tName]; ok {
			return fmt.Errorf("Duplicate tables with name %s", tName)
		}
		modelsM[tName] = model.TableDesc()
	}

	templateQStr := "CREATE TABLE IF NOT EXIST %s (%s)"

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, m := range models {

		qStr := fmt.Sprintf(templateQStr, m.TableName(), m.TableDesc())
		_, err := tx.Exec(qStr)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
