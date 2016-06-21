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
var ErrorNilQuitChanel = errors.New("Quit channel cannot be nil")
var ErrorBadUserID = errors.New("UserID cannot be empty")
var ErrorGCRunning = errors.New("Garbage collector is already running")

type Model struct {
	db *sql.DB

	gcRunning bool
	insCh     chan token
	delCh     chan string
}

func NewModel(db *sql.DB) (*Model, error) {

	if db == nil {
		return nil, helper.ErrorNilDB
	}

	iCh := make(chan token)
	dCh := make(chan string)

	m := &Model{
		db:    db,
		insCh: iCh,
		delCh: dCh,
	}

	return m, nil
}

func (m *Model) RunGarbageCollector(quitCh chan error) error {

	if quitCh == nil {
		return ErrorNilQuitChanel
	}

	smallest, err := m.GetSmallestExpiry()
	if err != nil {
		return err
	}

	go m.garbageCollect(smallest, quitCh)
	return nil
}

func (m *Model) TableName() string {
	return tableName
}

func (m *Model) PrimaryKeyField() string {
	return tableName + ".id"
}

func (m *Model) TableDesc() string {
	// TODO enforce fk integrity (userID)
	// CHECK https://github.com/cockroachdb/cockroach/issues/2132
	return fmt.Sprintf(`
		id	SERIAL		PRIMARY KEY,
		userID	INT		NOT NULL,
		devID	STRING		NOT NULL,
		token	STRING		UNIQUE NOT NULL,
		issued	TIMESTAMP	NOT NULL,
		expiry	TIMESTAMP	NOT NULL,
		INDEX userIDIndex (userID)
	`)
}

func (m *Model) Save(t token) (int, error) {

	qStr := fmt.Sprintf(`
		INSERT INTO %s (userID, devID, token, issued, expiry)
		VALUES ($1, $2, $3, $4, $5)
		 RETURNING id
		`, tableName)

	var tokenID int
	err := m.db.QueryRow(qStr, t.userID, t.devID, t.token,
		t.issued, t.expiry).Scan(&tokenID)
	if err != nil {
		return tokenID, err
	}

	t.id = tokenID
	if m.gcRunning {
		m.insCh <- t
	}
	return tokenID, nil
}

func (um *Model) Get(usrID int, tknStr string) (*token, error) {

	if usrID < 1 {
		return nil, ErrorBadUserID
	}

	if tknStr == "" {
		return nil, ErrorEmptyToken
	}

	qStr := fmt.Sprintf(`
		SELECT id, userID, devID, token, issued, expiry
		FROM %s
		WHERE userID = $1
		AND token = $2
	`, tableName)

	tkn := &token{}
	err := um.db.QueryRow(qStr, usrID, tknStr).Scan(
		&tkn.id, &tkn.userID, &tkn.devID, &tkn.token, &tkn.issued, &tkn.expiry,
	)
	if err != nil {
		if err.Error() == helper.NoResultsErrorStr {
			return nil, nil
		}
		return nil, err
	}

	tknI, err := um.ValidateExpiry(tkn)
	if err != nil {
		return nil, err
	}

	tkn, ok := tknI.(*token)
	if !ok {
		return nil, fmt.Errorf("Error handling token type got %T expected %T", tkn, &token{})
	}

	return tkn, nil
}

func (um *Model) Delete(tknStr string) (bool, error) {

	if tknStr == "" {
		return false, ErrorEmptyToken
	}

	qStr := fmt.Sprintf(`
		DELETE FROM %s
		WHERE token = $1
	`, tableName)

	r, err := um.db.Exec(qStr, tknStr)
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

	if um.gcRunning {
		um.delCh <- tknStr
	}
	return true, nil
}

func (um *Model) ValidateExpiry(tkn Token) (Token, error) {

	if time.Now().Before(tkn.Expiry()) {
		return tkn, nil
	}

	_, err := um.Delete(tkn.Token())
	if err != nil {
		return nil, fmt.Errorf("%s ...further %s while deleting the token",
			ErrorExpiredToken, err)
	}

	return nil, ErrorExpiredToken
}

func (um *Model) GetSmallestExpiry() (*token, error) {

	qStr := fmt.Sprintf(`
		SELECT id, userID, devID, token, issued, expiry
		FROM %s
		WHERE expiry = (SELECT MIN(expiry) FROM %s)
	`, tableName, tableName)

	t := &token{}
	err := um.db.QueryRow(qStr).Scan(
		&t.id, &t.userID, &t.devID, &t.token, &t.issued, &t.expiry,
	)

	if err != nil {
		if err.Error() == helper.NoResultsErrorStr {
			return nil, nil
		}
		return nil, err
	}

	return t, nil
}

// garbageCollect deletes all expired tokens.
// quitCh - sends an error value on quit (nil if no error occurred).
// garbageCollect is best run in a goroutine as it blocks in a
// forever loop until one of the following signals is received:
// 1. Token with earliest expiry time has just expired (garbage collect it).
// 2. A token has been inserted (recalculate token with earliest expiry time).
// 3. A token has been deleted (recalculate token with earliest expiry time).
func (um *Model) garbageCollect(smallest *token, quitCh chan error) {

	if quitCh == nil {
		return
	}

	if um.gcRunning {
		quitCh <- ErrorGCRunning
		return
	}

	um.gcRunning = true
	defer func() { um.gcRunning = false }()

	nextExp := longDuration
	if smallest != nil {
		nextExp = smallest.expiry.Sub(time.Now())
	}

	timer := time.NewTimer(nextExp)

loop:
	for {
		select {

		case <-timer.C:
			if smallest == nil {
				continue loop
			}
			// um.Delete() should send to delCh on success which will
			// trigger calculation of the next smallest
			_, err := um.Delete(smallest.token)
			if err != nil {
				quitCh <- err
				return
			}

		case inserted := <-um.insCh:
			if smallest == nil || smallest.expiry.After(inserted.expiry) {
				smallest = &inserted
			}

		case deleted := <-um.delCh:
			if deleted != smallest.token {
				continue
			}

			var err error
			smallest, err = um.GetSmallestExpiry()
			if err != nil {
				quitCh <- err
				return
			}

			nextExp = shortDuration
			if smallest != nil {
				nextExp = smallest.expiry.Sub(time.Now())
			}

			timer = time.NewTimer(nextExp)

		}
	}

	quitCh <- nil
}
