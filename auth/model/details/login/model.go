package login

import (
	"database/sql"
	"fmt"

	"errors"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
)

const (
	tableName = "details_login"
)

var ErrorEmptyUsersTable = errors.New("users table cannot be empty")

type Model struct {
	db               *sql.DB
	usersTable_IDCol string
}

func NewModel(db *sql.DB, usersTable_IDCol string) (*Model, error) {

	if db == nil {
		return nil, helper.ErrorNilDB
	}

	if usersTable_IDCol == "" {
		return nil, ErrorEmptyUsersTable
	}

	return &Model{db: db, usersTable_IDCol: usersTable_IDCol}, nil
}

func (m *Model) TableName() string {
	return tableName
}

func (m Model) PrimaryKeyField() string {
	return tableName + ".id"
}

func (m *Model) TableDesc() string {
	return fmt.Sprintf(`
		id		SERIAL		PRIMARY KEY,
		userID		INT 		NOT NULL CHECK (userID IN %s),
		date		TIMESTAMP	NOT NULL,
		forServiceID	STRING,
		ipAddress	STRING,
		referral	STRING,
		INDEX userLoginDateIndex (userID, date)
	`, m.usersTable_IDCol)
}

func (m *Model) Insert(ld LoginDetails) (int, error) {

	qStr := fmt.Sprintf(`
		INSERT INTO %s (userID, date, forServiceID, ipAddress, referral)
		VALUES ($1, $2, $3, $4, $5)
		 RETURNING id
		`, tableName)

	var detsID int
	err := m.db.QueryRow(qStr, ld.userID, ld.date,
		ld.forService, ld.ipAddress, ld.referral).Scan(&detsID)
	return detsID, err
}

func (m *Model) Get(userName, offset, count int) ([]*LoginDetails, error) {

	qStr := fmt.Sprintf(`
		SELECT id, userID, date, forServiceID, ipAddress, referral
		FROM %s
		WHERE userName = $1
		ORDER BY date DESC
		LIMIT $2, $3
	`, tableName)

	r, err := m.db.Query(qStr, userName, offset, count)
	if err != nil {
		return nil, err
	}

	dets := make([]*LoginDetails, count)
	d := &LoginDetails{}
	for i := 0; r.Next(); i++ {

		if i == count {
			return fmt.Errorf("Unexpected result, expected %d results currently at %d",
				count, i+1)
		}

		r.Scan(&d.id, &d.userID, &d.date, &d.forService,
			&d.ipAddress, &d.referral)
		dets[i] = d
	}

	return dets, r.Err()
}
