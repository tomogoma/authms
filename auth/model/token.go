package model

import (
	"errors"
	"fmt"
	"time"
	"github.com/tomogoma/go-commons/auth/token"
	"database/sql"
)

var ErrorEmptyToken = errors.New("Token string cannot be empty")
var ErrorExpiredToken = errors.New("Token has expired")
var ErrorNilQuitChanel = errors.New("Quit channel cannot be nil")
var ErrorNilLogger = errors.New("Logger cannot be nil")
var ErrorGCRunning = errors.New("Garbage collector is already running")
var ErrorInvalidToken = errors.New("Token is invalid")

type Logger interface {
	Info(interface{}, ...interface{})
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

func (m *Model) Save(t *token.Token) error {
	qStr := `
	INSERT INTO tokens (userID, devID, token, issued, expiry)
		VALUES ($1, $2, $3, $4, $5)
		 RETURNING id
	`
	err := m.db.QueryRow(qStr, t.UserID(), t.DevID(), t.Token(),
		t.Issued(), t.Expiry()).Scan(&t.Id)
	if err != nil {
		return err
	}
	if m.gcRunning {
		m.tokenSaveCh <- t
	}
	return nil
}

func (m *Model) Get(usrID int, devID, tknStr string) (*token.Token, error) {
	qStr := `
	SELECT id, userID, devID, token, issued, expiry
		FROM tokens
		WHERE userID = $1
		AND token = $2
		AND devID = $3
	`
	tkn := &token.Token{}
	err := m.db.QueryRow(qStr, usrID, tknStr, devID).Scan(
		&tkn.Id, &tkn.UsrID, &tkn.DvID, &tkn.TknStr, &tkn.IssueTime, &tkn.ExpiryTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return tkn, ErrorInvalidToken
		}
		return tkn, err
	}
	err = m.ValidateExpiry(tkn)
	return tkn, err
}

func (m *Model) Delete(tknStr string) (bool, error) {
	if tknStr == "" {
		return false, ErrorEmptyToken
	}
	qStr := `DELETE FROM tokens WHERE token = $1`
	r, err := m.db.Exec(qStr, tknStr)
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
	if m.gcRunning {
		m.tokenDelCh <- tknStr
	}
	return true, nil
}

func (um *Model) ValidateExpiry(tkn *token.Token) error {
	if time.Now().Before(tkn.Expiry()) {
		return nil
	}
	_, err := um.Delete(tkn.Token())
	if err != nil {
		return fmt.Errorf("%s ...further %s while deleting the token",
			ErrorExpiredToken, err)
	}
	return ErrorExpiredToken
}

func (um *Model) GetSmallestExpiry() (*token.Token, error) {
	qStr := `
	SELECT id, userID, devID, token, issued, expiry
		FROM tokens
		WHERE expiry = (SELECT MIN(expiry) FROM tokens)
	`
	t := &token.Token{}
	err := um.db.QueryRow(qStr).Scan(
		&t.Id, &t.UsrID, &t.DvID, &t.TknStr, &t.IssueTime, &t.ExpiryTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
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
func (um *Model) garbageCollect(smallest *token.Token, log Logger, quitCh chan error) {
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
	defer func() {
		um.gcRunning = false
	}()
	nextExp := token.ShortDuration
	if smallest != nil {
		nextExp = smallest.ExpiryTime.Sub(time.Now())
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
				nextExp = token.ShortDuration
				timer = time.NewTimer(nextExp)
				go um.concurrentDelete(smallest.TknStr, delErrCh)
				continue loop
			}
			var err error
			smallest, err = um.GetSmallestExpiry()
			if err != nil {
				quitCh <- err
				return
			}
			nextExp = token.ShortDuration
			if smallest != nil {
				nextExp = smallest.ExpiryTime.Sub(time.Now())
			}
			timer = time.NewTimer(nextExp)

		case inserted := <-um.tokenSaveCh:
			log.Info("Garbage collector - insertion occurred")
			if smallest != nil && smallest.Expiry().Before(inserted.Expiry()) {
				continue loop
			}

			smallest = inserted
			nextExp = smallest.Expiry().Sub(time.Now())
			timer = time.NewTimer(nextExp)

		case deleted := <-um.tokenDelCh:
			log.Info("Garbage collector - deletion occured")
			if smallest != nil && deleted != smallest.Token() {
				continue loop
			}

			var err error
			smallest, err = um.GetSmallestExpiry()
			if err != nil {
				quitCh <- err
				return
			}

			nextExp = token.ShortDuration
			if smallest != nil {
				nextExp = smallest.Expiry().Sub(time.Now())
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
