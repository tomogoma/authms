package token

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/tomogoma/authms/auth/model/helper"
)

var ErrorEmptyToken = errors.New("Token string cannot be empty")
var ErrorExpiredToken = errors.New("Token has expired")
var ErrorNilQuitChanel = errors.New("Quit channel cannot be nil")
var ErrorNilLogger = errors.New("Logger cannot be nil")
var ErrorBadUserID = errors.New("UserID cannot be empty")
var ErrorGCRunning = errors.New("Garbage collector is already running")
var ErrorInvalidToken = errors.New("Token is invalid")

type Logger interface {
	Info(interface{}, ...interface{})
}

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

func (m *Model) RunGarbageCollector(quitCh chan error, lg Logger) error {

	if quitCh == nil {
		return ErrorNilQuitChanel
	}

	if lg == nil {
		return ErrorNilLogger
	}

	smallest, err := m.GetSmallestExpiry()
	if err != nil {
		return err
	}

	go m.garbageCollect(smallest, lg, quitCh)
	return nil
}

func (m *Model) Save(t token) (int, error) {

	qStr := `
	INSERT INTO tokens (userID, devID, token, issued, expiry)
		VALUES ($1, $2, $3, $4, $5)
		 RETURNING id
	`

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

func (um *Model) Get(usrID int, devID, tknStr string) (*token, error) {

	qStr := `
	SELECT id, userID, devID, token, issued, expiry
		FROM tokens
		WHERE userID = $1
		AND token = $2
		AND devID = $3
	`

	tkn := &token{}
	err := um.db.QueryRow(qStr, usrID, tknStr, devID).Scan(
		&tkn.id, &tkn.userID, &tkn.devID, &tkn.token, &tkn.issued, &tkn.expiry,
	)
	if err != nil {
		if err.Error() == helper.NoResultsErrorStr {
			return nil, ErrorInvalidToken
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

	qStr := `DELETE FROM tokens WHERE token = $1`

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

	qStr := `
	SELECT id, userID, devID, token, issued, expiry
		FROM tokens
		WHERE expiry = (SELECT MIN(expiry) FROM tokens)
	`

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

// um.concurrentDelete() deletes a token with token string tknstr and
// sends to delCh on success or delErrCh on failure
func (um *Model) concurrentDelete(tknStr string, errCh chan error) {

	if errCh == nil {
		return
	}

	_, err := um.Delete(tknStr)
	if err != nil {
		errCh <- err
	}
}

// garbageCollect deletes all expired tokens.
// quitCh - sends an error value on quit (nil if no error occurred).
// garbageCollect is best run in a goroutine as it blocks in a
// forever loop until one of the following signals is received:
// 1. Token with earliest expiry time has just expired (garbage collect it).
// 2. A token has been inserted (recalculate token with earliest expiry time).
// 3. A token has been deleted (recalculate token with earliest expiry time).
func (um *Model) garbageCollect(smallest *token, log Logger, quitCh chan error) {

	if quitCh == nil {
		return
	}

	if log == nil {
		quitCh <- ErrorNilLogger
	}

	if um.gcRunning {
		quitCh <- ErrorGCRunning
		return
	}

	um.gcRunning = true
	defer func() { um.gcRunning = false }()

	nextExp := shortDuration
	if smallest != nil {
		nextExp = smallest.expiry.Sub(time.Now())
	}

	timer := time.NewTimer(nextExp)
	delErrCh := make(chan error)

loop:
	for {
		log.Info("Garbage collector - next expiry in less than %v", nextExp)
		select {

		case <-timer.C:
			log.Info("Garbage collector - expiry occured")
			if smallest != nil {
				nextExp = shortDuration
				timer = time.NewTimer(nextExp)
				go um.concurrentDelete(smallest.token, delErrCh)
				continue loop
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

		case inserted := <-um.insCh:
			log.Info("Garbage collector - insertion occurred")
			if smallest != nil && smallest.expiry.Before(inserted.expiry) {
				continue loop
			}

			smallest = &inserted
			nextExp = smallest.expiry.Sub(time.Now())
			timer = time.NewTimer(nextExp)

		case deleted := <-um.delCh:
			log.Info("Garbage collector - deletion occured")
			if smallest != nil && deleted != smallest.token {
				continue loop
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

		case delErr := <-delErrCh:
			log.Info("Garbage collector - deletion error: %v", delErr)
			if delErr == nil {
				continue loop
			}
			quitCh <- delErr
			return

		}
	}

	quitCh <- nil
}
