package user

import (
	"database/sql"
	"fmt"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
)

const (
	tableName = "users"
)

type Model struct {
	db *sql.DB
}

func NewModel(db *sql.DB) (*Model, error) {

	if db == nil {
		return nil, helper.ErrorNilDB
	}

	return &Model{db: db}, nil
}

func (m Model) TableName() string {
	return tableName
}

func (m Model) PrimaryKeyField() string {
	return tableName + ".id"
}

func (m Model) TableDesc() string {
	return `
		id		SERIAL	PRIMARY KEY,
		userName	STRING	UNIQUE,
		password	BYTES	NOT NULL,
		firstName	STRING,
		middleName	STRING,
		lastName	STRING
	`
}

func (m Model) Save(u user) (int, error) {

	qStr := fmt.Sprintf(`
		INSERT INTO %s (userName, firstName, middleName, lastName, password)
		VALUES ($1, $2, $3, $4, $5)
		 RETURNING id
		`, tableName)

	var userid int
	err := m.db.QueryRow(qStr, u.userName, u.firstName,
		u.middleName, u.lastName, u.password).Scan(&userid)
	return userid, err
}

func (m Model) GetByID(userID int) (*user, error) {

	qStr := fmt.Sprintf(`
		SELECT id, userName, password, firstName, middleName, lastName
		FROM %s
		WHERE id = $1
	`, tableName)

	usr := &user{}
	err := m.db.QueryRow(qStr, userID).Scan(
		&usr.id, &usr.userName, &usr.password,
		&usr.firstName, &usr.middleName, &usr.lastName,
	)

	if err != nil {
		if err.Error() != helper.NoResultsErrorStr {
			return nil, err
		}
		return nil, nil
	}

	return usr, err
}

func (m Model) Get(uName, pass string, hashF HashFunc) (*user, error) {

	qStr := fmt.Sprintf(`
		SELECT id, userName, password, firstName, middleName, lastName
		FROM %s
		WHERE userName = $1
		AND password = $2
	`, tableName)

	passHB, err := hashF(pass)
	if err != nil {
		return nil, err
	}

	usr := &user{}
	err = m.db.QueryRow(qStr, uName, passHB).Scan(
		&usr.id, &usr.userName, &usr.password,
		&usr.firstName, &usr.middleName, &usr.lastName,
	)

	if err != nil {
		if err.Error() != helper.NoResultsErrorStr {
			return nil, err
		}
		return nil, nil
	}

	return usr, err
}
