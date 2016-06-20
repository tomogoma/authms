package token

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
)

const (
	tableName = "tokens"
)

var ErrorEmptyToken = errors.New("Token string cannot be empty")
var ErrorExpiredToken = errors.New("Token has expired")
var ErrorEmptyUsersTable = errors.New("users table cannot be empty")
var ErrorEmptyUserID = errors.New("UserID cannot be empty")

type Model struct {
	db               *sql.DB
	usersTable_IDCol string

	insCh chan Token
	delCh chan string
}

func NewModel(db *sql.DB, usersTable_IDCol string, quitCh chan error) (*Model, error) {

	if db == nil {
		return nil, helper.ErrorNilDB
	}

	if usersTable_IDCol == "" {
		return nil, ErrorEmptyUsersTable
	}

	iCh := make(chan Token)
	dCh := make(chan string)

	m := &Model{
		db:               db,
		usersTable_IDCol: usersTable_IDCol,
		insCh:            iCh,
		delCh:            dCh,
	}

	go m.garbageCollect(quitCh)

	return m, nil
}

func (m *Model) TableName() string {
	return tableName
}

func (m Model) PrimaryKeyField() string {
	return tableName + ".id"
}

func (m *Model) TableDesc() string {
	return fmt.Sprintf(`
		id	SERIAL		PRIMARY KEY,
		userID	INT		NOT NULL CHECK (userID IN %s),
		devID	STRING		NOT NULL,
		token	STRING		UNIQUE NOT NULL,
		issue	TIMESTAMP	NOT NULL,
		expiry	TIMESTAMP	NOT NULL,
		INDEX userIDIndex (userID)
	`, m.usersTable_IDCol)
}

func (m *Model) Save(t Token) (int, error) {

	qStr := fmt.Sprintf(`
		INSERT INTO %s (userID, devID, token, issue, expiry)
		VALUES ($1, $2, $3, $4, $5)
		 RETURNING id
		`, tableName)

	var tokenID int
	err := m.db.QueryRow(qStr, t.userID, t.devID, t.token,
		t.issue, t.expiry).Scan(&tokenID)
	if err != nil {
		return tokenID, err
	}

	t.id = tokenID
	m.insCh <- t
	return tokenID, nil
}

func (um Model) Get(usrID int, token string) (*Token, error) {

	if usrID < 1 {
		return nil, ErrorEmptyUserID
	}

	if token == "" {
		return nil, ErrorEmptyToken
	}

	qStr := fmt.Sprintf(`
		SELECT id, userID, devID, token, issue, expiry
		FROM %s
		WHERE userID = $1
		AND token = $2
	`, tableName)

	var t *Token
	err := um.db.QueryRow(qStr, usrID, token).Scan(
		&t.id, &t.userID, &t.devID, &t.token, &t.issue, &t.expiry,
	)
	if err != nil {
		return nil, err
	}

	if time.Now().After(t.expiry) {

		ok, err := um.Delete(token)
		if err != nil {
			return nil, fmt.Errorf("%s ...further %s while deleting the token",
				ErrorExpiredToken, err)
		}

		if !ok {
			return nil, fmt.Errorf("%s ...further, token was not deleted successfully",
				ErrorExpiredToken)
		}

		return nil, ErrorExpiredToken
	}

	return t, nil
}

func (um Model) Delete(token string) (bool, error) {

	qStr := fmt.Sprintf(`
		DELETE FROM %s
		WHERE token = $1
	`, tableName)

	r, err := um.db.Exec(qStr, token)
	if err != nil {
		return false, err
	}

	i, err := r.RowsAffected()
	if err != nil {
		return false, err
	}

	if i > 1 {
		return false, fmt.Errorf("Expected 1 rows affected but got %d", i)
	}

	if i == 0 {
		return false, nil
	}

	um.delCh <- token
	return true, nil
}

func (um Model) garbageCollect(quitCh chan error) {

	smallest, err := um.getSmallestExpiry()
	if err != nil {
		quitCh <- err
		return
	}

	nextExp := smallest.expiry.Sub(time.Now())
	timer := time.NewTimer(nextExp)

	for {
		select {

		case <-timer.C:
			_, err := um.Delete(smallest.token)
			if err != nil {
				quitCh <- err
				return
			}

		case inserted := <-um.insCh:
			if smallest.expiry.After(inserted.expiry) {
				smallest = &inserted
			}

		case deleted := <-um.delCh:
			if deleted != smallest.token {
				continue
			}

			smallest, err = um.getSmallestExpiry()
			if err != nil {
				quitCh <- err
				return
			}

		}
	}

	quitCh <- nil
}

func (um Model) getSmallestExpiry() (*Token, error) {

	qStr := fmt.Sprintf(`
		SELECT id, userID, devID, token, issue, expiry
		FROM %s
		WHERE expiry = (SELECT MIN(date) FROM %s)
	`, tableName, tableName)

	var t *Token
	err := um.db.QueryRow(qStr).Scan(
		&t.id, &t.userID, &t.devID, &t.token, &t.issue, &t.expiry,
	)
	if err != nil {
		return nil, err
	}

	return t, nil
}
