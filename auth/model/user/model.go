package user

import (
	"database/sql"

	"errors"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/lib/pq"
	"github.com/tomogoma/authms/auth/model/helper"
)

var ErrorPasswordMismatch = errors.New("username/password combo mismatch")
var ErrorUserExists = errors.New("A user with the provided username already exists")
var ErrorEmailExists = errors.New("A user with the provided email already exists")
var ErrorAppIDExists = errors.New("A user with the provided app ID for the provided app name already exists")
var ErrorPhoneExists = errors.New("A user with the provided phone already exists")
var ErrorModelCorruptedOnEmptyPassword = errors.New("The model contained an empty password value and is probably corrupt")

type Model struct {
	db *sql.DB
}

func NewModel(db *sql.DB) (*Model, error) {

	if db == nil {
		return nil, helper.ErrorNilDB
	}

	return &Model{db: db}, nil
}

func (m Model) Save(u user) (*user, error) {
	err := crdb.ExecuteTx(m.db, func(tx *sql.Tx) error {
		userqStr := `INSERT INTO users (password, createDate)
	 		VALUES ($1, CURRENT_TIMESTAMP()) RETURNING id`
		err := tx.QueryRow(userqStr, u.password).Scan(&u.id)
		if err = processError(err, err); err != nil {
			return err
		}
		if valIsFilled(u.email) {
			emailqStr := `INSERT INTO emails (userID, email, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(emailqStr, u.id, u.email.value)
			if err = processError(err, ErrorEmailExists); err != nil {
				return err
			}
		}
		if u.userName != "" {
			usrnmqStr := `INSERT INTO userNames (userID, userName, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(usrnmqStr, u.id, u.userName)
			if err = processError(err, ErrorUserExists); err != nil {
				return err
			}
		}
		if valIsFilled(u.phone) {
			phoneqStr := `INSERT INTO phones (userID, phone, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(phoneqStr, u.id, u.phone.value)
			if err = processError(err, ErrorPhoneExists); err != nil {
				return err
			}
		}
		if appIsFilled(u.app) {
			extqStr := `INSERT INTO appUserIDs (userID, appUserID, appName, createDate)
	 		VALUES ($1, $2, $3, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(extqStr, u.id, u.app.userID, u.app.name)
			if err = processError(err, ErrorAppIDExists); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m Model) Get(userID int) (*user, error) {
	userQ := `SELECT id FROM users WHERE id = $1`
	usr := &user{}
	err := m.db.QueryRow(userQ, userID).Scan(&usr.id)
	if err != nil {
		return nil, err
	}
	return usr, err
}

func (m Model) GetByUserName(uName, pass string, hashF ValidatePassFunc) (*user, error) {
	usr := &user{}
	unameQ := `SELECT userID, userName FROM userNames WHERE userName = $1`
	err := m.db.QueryRow(unameQ, uName).Scan(&usr.id, &usr.userName)
	if err != nil {
		if err.Error() != helper.NoResultsErrorStr {
			return nil, err
		}
		return nil, ErrorPasswordMismatch
	}
	if err = m.validatePassword(usr.id, pass, hashF); err != nil {
		return nil, err
	}
	return usr, nil
}

func (m Model) GetByPhone(phone, pass string, hashF ValidatePassFunc) (*user, error) {
	usr := &user{phone: &value{}}
	query := `SELECT userID, phone, validated FROM phones WHERE phone = $1`
	err := m.db.QueryRow(query, phone).
		Scan(&usr.id, &usr.phone.value, &usr.phone.validated)
	if err != nil {
		if err.Error() != helper.NoResultsErrorStr {
			return nil, err
		}
		return nil, ErrorPasswordMismatch
	}
	if err = m.validatePassword(usr.id, pass, hashF); err != nil {
		return nil, err
	}
	return usr, nil
}

func (m Model) GetByEmail(email, pass string, hashF ValidatePassFunc) (*user, error) {
	usr := &user{email: &value{}}
	query := `SELECT userID, email, validated FROM emails WHERE email = $1`
	err := m.db.QueryRow(query, email).
		Scan(&usr.id, &usr.email.value, &usr.email.validated)
	if err != nil {
		if err.Error() != helper.NoResultsErrorStr {
			return nil, err
		}
		return nil, ErrorPasswordMismatch
	}
	if err = m.validatePassword(usr.id, pass, hashF); err != nil {
		return nil, err
	}
	return usr, nil
}

func (m Model) GetByAppUserID(appName, appUserID string) (*user, error) {
	usr := &user{app: &app{}}
	query := `SELECT userID, appName, appUserID, validated FROM appUserIDs
	 		WHERE appName = $1 AND appUserID = $2`
	err := m.db.QueryRow(query, appName, appUserID).
		Scan(&usr.id, &usr.app.name, &usr.app.userID, &usr.app.validated)
	if err != nil {
		if err.Error() != helper.NoResultsErrorStr {
			return nil, err
		}
		return nil, ErrorPasswordMismatch
	}
	return usr, nil
}

func (m Model) validatePassword(id int, password string, hashF ValidatePassFunc) error {
	userQ := `SELECT password FROM users WHERE id = $1`
	var dbPassword []byte
	err := m.db.QueryRow(userQ, id).Scan(&dbPassword)
	if err != nil {
		return err
	}
	if len(dbPassword) == 0 {
		return ErrorModelCorruptedOnEmptyPassword
	}
	if !hashF(password, dbPassword) {
		return ErrorPasswordMismatch
	}
	return err
}

func processError(receivedErr, existsErr error) error {
	if receivedErr != nil {
		if pqErr, ok := receivedErr.(*pq.Error); ok && pqErr.Code == "23505" {
			receivedErr = existsErr
		}
		return receivedErr
	}
	return nil
}
