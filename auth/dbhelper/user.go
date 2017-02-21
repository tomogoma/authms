package dbhelper

import (
	"github.com/cockroachdb/cockroach-go/crdb"
	"fmt"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/errors"
	"database/sql"
	"github.com/lib/pq"
	"github.com/tomogoma/go-commons/auth/token"
	"log"
)

type PasswordGenerator interface {
	SecureRandomString(length int) ([]byte, error)
}

type TokenValidator interface {
	Validate(tokenStr string) (*token.Token, error)
}

type Hasher interface {
	Hash(pass string) ([]byte, error)
	CompareHash(pass string, passHB []byte) bool
}

var ErrorPasswordMismatch = errors.New("username/password combo mismatch")
var ErrorUserExists = errors.New("A user with some of the provided details already exists")
var ErrorEmailExists = errors.New("A user with the provided email already exists")
var ErrorAppIDExists = errors.New("A user with the provided app ID for the provided app name already exists")
var ErrorPhoneExists = errors.New("A user with the provided phone already exists")
var ErrorInvalidOAuth = errors.New("the oauth provided was invalid")
var ErrorEmptyEmail = errors.New("email was empty")
var ErrorEmptyPhone = errors.New("phone was empty")
var ErrorEmptyUserName = errors.New("username was empty")
var ErrorEmptyPassword = errors.New("password cannot be empty")
var ErrorModelCorruptedOnEmptyPassword = errors.New("The model contained an empty password value and is probably corrupt")

func (m *DBHelper) SaveUser(u *authms.User) error {
	if u == nil {
		return errors.New("user was nil")
	}
	passHB, err := m.getPasswordHash(u)
	if err != nil {
		return err
	}
	err = crdb.ExecuteTx(m.db, func(tx *sql.Tx) error {
		userqStr := `INSERT INTO users (password, createDate)
	 		VALUES ($1, CURRENT_TIMESTAMP()) RETURNING id`
		err := tx.QueryRow(userqStr, passHB).Scan(&u.ID)
		if err != nil {
			return err
		}
		if HasValue(u.Email) {
			emailqStr := `INSERT INTO emails (userID, email, validated, createDate)
		 		VALUES ($1, $2, $3, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(emailqStr, u.ID, u.Email.Value, u.Email.Verified)
			if err != nil {
				return err
			}
		}
		if u.UserName != "" {
			usrnmqStr := `INSERT INTO userNames (userID, userName, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(usrnmqStr, u.ID, u.UserName)
			if err != nil {
				return err
			}
		}
		if HasValue(u.Phone) {
			phoneqStr := `INSERT INTO phones (userID, phone, validated, createDate)
		 		VALUES ($1, $2, $3, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(phoneqStr, u.ID, u.Phone.Value, u.Phone.Verified)
			if err != nil {
				return err
			}
		}
		for _, oAuth := range u.OAuths {
			extqStr := `INSERT INTO appUserIDs (userID, appUserID, appName, validated, createDate)
	 		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(extqStr, u.ID, oAuth.AppUserID, oAuth.AppName, oAuth.Verified)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return extractDuplicateError(err)
}

func (m *DBHelper) GetByUserName(uName, pass string) (*authms.User, error) {
	where := "usernames.userName = $1"
	usr, err := m.get(where, uName)
	return m.validateFetchedUser(usr, err, pass)
}

func (m *DBHelper) GetByPhone(phone, pass string) (*authms.User, error) {
	where := `phones.phone = $1`
	usr, err := m.get(where, phone)
	return m.validateFetchedUser(usr, err, pass)
}

func (m *DBHelper) GetByEmail(email, pass string) (*authms.User, error) {
	where := `emails.email = $1`
	usr, err := m.get(where, email)
	return m.validateFetchedUser(usr, err, pass)
}

func (m *DBHelper) GetByAppUserID(appName, appUserID string) (*authms.User, error) {
	usr := &authms.User{}
	query := `SELECT userID FROM appUserIDs WHERE appName = $1 AND appUserID = $2`
	err := m.db.QueryRow(query, appName, appUserID).Scan(&usr.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			return usr, err
		}
		return usr, ErrorPasswordMismatch
	}
	where := "users.id=$1"
	return m.get(where, usr.ID)
}

func (m *DBHelper) UpdateUserName(userID int64, newUserName string) error {
	if newUserName == "" {
		return errors.New("the userName provided was invlaid")
	}
	q := `SELECT COUNT(id) FROM userNames WHERE userID=$1`
	var count int
	if err := m.db.QueryRow(q, userID).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has usernae: %s", err)
	}
	if (count == 0) {
		q = `INSERT INTO userNames (userID, userName, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
		rslt, err := m.db.Exec(q, userID, newUserName)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE userNames
		 	SET userName=$1, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$2`
	rslt, err := m.db.Exec(q, newUserName, userID)
	return checkRowsAffected(rslt, err, 1)
}

func (m *DBHelper) UpdateAppUserID(userID int64, new *authms.OAuth) error {
	if new == nil {
		return errors.New("new OAuth was nil")
	}
	q := `SELECT COUNT(id) FROM appUserIDs WHERE userID=$1 AND appName=$2`
	var count int
	if err := m.db.QueryRow(q, userID, new.AppName).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has email: %s", err)
	}
	if (count == 0) {
		q = `INSERT INTO appUserIDs (userID, appUserID, appName, validated, createDate)
	 		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP())`
		rslt, err := m.db.Exec(q, userID, new.AppUserID, new.AppName, new.Verified)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE appUserIDs
		 	SET appUserID=$1, validated=$2, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$3 AND appName=$4`
	rslt, err := m.db.Exec(q, new.AppUserID, new.Verified, userID, new.AppName)
	return checkRowsAffected(rslt, err, 1)
}

func (m *DBHelper) UpdateEmail(userID int64, newEmail *authms.Value) error {
	if !HasValue(newEmail) {
		return errors.New("the email provided was invlaid")
	}
	q := `SELECT COUNT(id) FROM emails WHERE userID=$1`
	var count int
	if err := m.db.QueryRow(q, userID).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has email: %s", err)
	}
	if (count == 0) {
		q = `INSERT INTO emails (userID, email, validated, createDate)
		 		VALUES ($1, $2, $3, CURRENT_TIMESTAMP())`
		rslt, err := m.db.Exec(q, userID, newEmail.Value, newEmail.Verified)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE emails
		 	SET email=$1, validated=$2, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$3`
	rslt, err := m.db.Exec(q, newEmail.Value, newEmail.Verified, userID)
	return checkRowsAffected(rslt, err, 1)
}

func (m *DBHelper) UpdatePhone(userID int64, newPhone *authms.Value) error {
	if !HasValue(newPhone) {
		return errors.New("the phone provided was invlaid")
	}
	q := `SELECT COUNT(id) FROM phones WHERE userID=$1`
	var count int
	if err := m.db.QueryRow(q, userID).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has phone: %s", err)
	}
	if (count == 0) {
		q = `INSERT INTO phones (userID, phone, validated, createDate)
		 		VALUES ($1, $2, $3, CURRENT_TIMESTAMP())`
		rslt, err := m.db.Exec(q, userID, newPhone.Value, newPhone.Verified)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE phones
		 	SET phone=$1, validated=$2, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$3`
	rslt, err := m.db.Exec(q, newPhone.Value, newPhone.Verified, userID)
	return checkRowsAffected(rslt, err, 1)
}

func (m *DBHelper) UpdatePassword(userID int64, oldPass, newPassword string) error {
	q := `SELECT password FROM users WHERE id=$1`
	var actPassHB []byte
	err := m.db.QueryRow(q, userID).Scan(&actPassHB)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("not found")
			return ErrorPasswordMismatch
		}
		return err
	}
	if ! m.hasher.CompareHash(oldPass, actPassHB) {
		return ErrorPasswordMismatch
	}
	passHB, err := m.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	q = `UPDATE users
			SET password=$1, updateDate=CURRENT_TIMESTAMP()
	 		WHERE id=$2`
	rslt, err := m.db.Exec(q, passHB, userID)
	return checkRowsAffected(rslt, err, 1)
}

func (m *DBHelper) validateFetchedUser(usr *authms.User, getErr error, pass string) (
*authms.User, error) {
	if getErr != nil {
		return usr, getErr
	}
	if getErr = m.validatePassword(usr.ID, pass); getErr != nil {
		return usr, getErr
	}
	return usr, nil
}

func (m *DBHelper) get(where string, whereArgs... interface{}) (*authms.User, error) {
	usr := &authms.User{
		Email: &authms.Value{},
		Phone: &authms.Value{},
		OAuths: make(map[string]*authms.OAuth),
	}
	query := `
	SELECT
		users.id, userNames.userName, phones.phone, phones.validated,
		emails.email, emails.validated
		FROM users
		LEFT JOIN userNames ON users.id=userNames.userID
		LEFT JOIN phones ON users.id=phones.userID
		LEFT JOIN emails ON users.id=emails.userID
		WHERE `
	query = fmt.Sprintf("%s%s", query, where)
	var dbUserName, dbPhone, dbEmail sql.NullString
	var dbPhoneValidated, dbEmailValidated sql.NullBool
	err := m.db.QueryRow(query, whereArgs...).Scan(&usr.ID, &dbUserName,
		&dbPhone, &dbPhoneValidated, &dbEmail, &dbEmailValidated)
	usr.UserName = dbUserName.String
	usr.Phone.Value = dbPhone.String
	usr.Phone.Verified = dbPhoneValidated.Bool
	usr.Email.Value = dbEmail.String
	usr.Email.Verified = dbEmailValidated.Bool
	if err != nil {
		if err != sql.ErrNoRows {
			return usr, err
		}
		return usr, ErrorPasswordMismatch
	}
	query = `SELECT appUserID, appName, validated FROM appUserIDs WHERE userID=$1`
	rslt, err := m.db.Query(query, usr.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return usr, nil
		}
		return usr, err
	}
	for rslt.Next() {
		app := new(authms.OAuth)
		if err := rslt.Scan(&app.AppUserID, &app.AppName, &app.Verified); err != nil {
			return usr, err
		}
		usr.OAuths[app.AppName] = app
	}
	if err = rslt.Close(); err != nil {
		return usr, err
	}
	return usr, nil
}

func (m *DBHelper) validatePassword(id int64, password string) error {
	userQ := `SELECT password FROM users WHERE id = $1`
	var dbPassword []byte
	err := m.db.QueryRow(userQ, id).Scan(&dbPassword)
	if err != nil {
		return err
	}
	if len(dbPassword) == 0 {
		return ErrorModelCorruptedOnEmptyPassword
	}
	if !m.hasher.CompareHash(password, dbPassword) {
		return ErrorPasswordMismatch
	}
	return err
}

func (m *DBHelper) getPasswordHash(u *authms.User) ([]byte, error) {
	passStr := u.Password
	if passStr == "" {
		passB, err := m.gen.SecureRandomString(36)
		if err != nil {
			return nil, errors.Newf("error generating password: %v",
				err)
		}
		passStr = string(passB)
	}
	return m.hasher.Hash(passStr)
}

func extractDuplicateError(err error) error {
	if err == nil {
		return nil
	}
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return ErrorUserExists
	}
	return err
}

func HasValue(v *authms.Value) bool {
	return v != nil && v.Value != ""
}
