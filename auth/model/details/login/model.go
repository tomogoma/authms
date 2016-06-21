package login

import (
	"database/sql"
	"fmt"

	"time"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
)

const (
	tableName  = "details_login"
	dateFormat = time.RFC3339 // "2006-01-02T15:04:05Z07:00"
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

func (m *Model) TableName() string {
	return tableName
}

func (m Model) PrimaryKeyField() string {
	return tableName + ".id"
}

func (m *Model) TableDesc() string {
	// TODO enforce fk integrity (userID)
	// CHECK https://github.com/cockroachdb/cockroach/issues/2132
	return fmt.Sprintf(`
		id		SERIAL		PRIMARY KEY,
		userID		INT 		NOT NULL,
		date		TIMESTAMP	NOT NULL,
		forServiceID	STRING,
		ipAddress	STRING,
		referral	STRING,
		INDEX userLoginDateIndex (userID, date)
	`)
}

func (m *Model) Save(ld loginDetails) (int, error) {

	qStr := fmt.Sprintf(`
		INSERT INTO %s (userID, date, forServiceID, ipAddress, referral)
		VALUES ($1, $2, $3, $4, $5)
		 RETURNING id
		`, tableName)

	var detsID int
	date := ld.date.Format(dateFormat)
	err := m.db.QueryRow(qStr, ld.userID, date,
		ld.forService, ld.ipAddress, ld.referral).Scan(&detsID)
	return detsID, err
}

func (m *Model) Get(userID, offset, count int) ([]*loginDetails, error) {

	qStr := fmt.Sprintf(`
		SELECT id, userID, date, forServiceID, ipAddress, referral
		FROM %s
		WHERE userID = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`, tableName)

	r, err := m.db.Query(qStr, userID, count, offset)
	if err != nil {
		return nil, err
	}

	dets := make([]*loginDetails, 0)
	for r.Next() {

		d := &loginDetails{}
		var tmStmp string
		r.Scan(&d.id, &d.userID, &tmStmp, &d.forService, &d.ipAddress, &d.referral)
		d.date, err = time.Parse(dateFormat, tmStmp)
		if err != nil {
			return nil, err
		}
		dets = append(dets, d)
	}

	return dets, r.Err()
}
