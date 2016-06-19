package model

import (
	"database/sql"
	"fmt"

	"bitbucket.org/alkira/contactsms/kazoo/errors"
	_ "github.com/lib/pq"
)

const (
	driverName = "postgres"
)

var ErrorNilDSNFormatter = errors.New("DSNFormatter cannot be nil")

type Model interface {
	TableName() string
	TableDesc() string
}

type DSNFormatter interface {
	FormatDSN() string
}

type DSN struct {
	UName    string
	Password string
	Host     string
	DB       string
}

func (d DSN) FormatDSN() string {

	if d.UName != "" {
		if d.Password != "" {
			return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=verify-full",
				d.UName, d.Password, d.Host, d.DB,
			)
		}
		return fmt.Sprintf("postgres://%s@%s/%s?sslmode=verify-full",
			d.UName, d.Host, d.DB,
		)
	}

	return fmt.Sprintf("postgres://%s/%s?sslmode=verify-full", d.Host, d.DB)
}

func New(dsnF DSNFormatter) (*sql.DB, error) {

	if dsnF == nil {
		return ErrorNilDSNFormatter
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
