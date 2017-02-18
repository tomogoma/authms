package user

import (
	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/tomogoma/authms/auth/model/helper"
	"fmt"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/errors"
	"github.com/tomogoma/go-commons/database/cockroach"
	"database/sql"
	"github.com/lib/pq"
	"github.com/tomogoma/go-commons/auth/token"
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

type Model struct {
	db     *sql.DB
	hasher Hasher
	gen    PasswordGenerator
	token  TokenValidator
}

var ErrorPasswordMismatch = errors.New("username/password combo mismatch")
var ErrorNilHashFunc = errors.New("HashFunc cannot be nil")
var ErrorUserExists = errors.New("A user with some of the provided details already exists")
var ErrorEmailExists = errors.New("A user with the provided email already exists")
var ErrorAppIDExists = errors.New("A user with the provided app ID for the provided app name already exists")
var ErrorPhoneExists = errors.New("A user with the provided phone already exists")
var ErrorInvalidOAuth = errors.New("the oauth provided was invalid")
var ErrorEmptyEmail = errors.New("email was empty")
var ErrorEmptyPhone = errors.New("phone was empty")
var ErrorEmptyUserName = errors.New("username was empty")
var ErrorEmptyPassword = errors.New("password cannot be empty")
var ErrorNilPasswordGenerator = errors.New("password generator was nil")
var ErrorNilTokenValidator = errors.New("token validator was nil")
var ErrorModelCorruptedOnEmptyPassword = errors.New("The model contained an empty password value and is probably corrupt")

func NewModel(dsnF cockroach.DSNFormatter, pg PasswordGenerator, h Hasher, tv TokenValidator) (*Model, error) {
	if h == nil {
		return nil, ErrorNilHashFunc
	}
	if pg == nil {
		return nil, ErrorNilPasswordGenerator
	}
	if tv == nil {
		return nil, ErrorNilTokenValidator
	}
	db, err := cockroach.DBConn(dsnF)
	if err != nil {
		return nil, errors.Newf("error connecting to db: %s", err)
	}
	if err := cockroach.InstantiateDB(db, dsnF.DBName(), users, usernames, emails,
		phones, appUserIDs); err != nil {
		return nil, errors.Newf("error instantiating db: %s", err)
	}
	return &Model{db: db, gen: pg, hasher: h, token: tv}, nil
}

func (m Model) Save(u *authms.User) error {
	if err := validateUser(u); err != nil {
		return err
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
			emailqStr := `INSERT INTO emails (userID, email, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(emailqStr, u.ID, u.Email.Value)
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
			phoneqStr := `INSERT INTO phones (userID, phone, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(phoneqStr, u.ID, u.Phone.Value)
			if err != nil {
				return err
			}
		}
		if u.OAuth != nil {
			extqStr := `INSERT INTO appUserIDs (userID, appUserID, appName, createDate)
	 		VALUES ($1, $2, $3, CURRENT_TIMESTAMP())`
			_, err = tx.Exec(extqStr, u.ID, u.OAuth.AppUserID, u.OAuth.AppName)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return extractDuplicateError(err)
}

func (m Model) GetByUserName(uName, pass string) (*authms.User, error) {
	where := "usernames.userName = $1"
	usr, err := m.get(where, uName)
	if err != nil {
		return usr, err
	}
	if err = m.validatePassword(usr.ID, pass); err != nil {
		return usr, err
	}
	return usr, nil
}

func (m Model) GetByPhone(phone, pass string) (*authms.User, error) {
	where := `phones.phone = $1`
	usr, err := m.get(where, phone)
	if err != nil {
		return usr, err
	}
	if err = m.validatePassword(usr.ID, pass); err != nil {
		return usr, err
	}
	return usr, nil
}

func (m Model) GetByEmail(email, pass string) (*authms.User, error) {
	where := `emails.email = $1`
	usr, err := m.get(where, email)
	if err != nil {
		return usr, err
	}
	if err = m.validatePassword(usr.ID, pass); err != nil {
		return usr, err
	}
	return usr, err
}

func (m Model) GetByAppUserID(appName, appUserID, appToken string) (*authms.User, error) {
	// TODO validate appToken
	usr := &authms.User{}
	query := `SELECT userID FROM appUserIDs WHERE appName = $1 AND appUserID = $2`
	err := m.db.QueryRow(query, appName, appUserID).Scan(&usr.ID)
	if err != nil {
		if err.Error() != helper.NoResultsErrorStr {
			return usr, err
		}
		return usr, ErrorPasswordMismatch
	}
	where := "users.id=$1"
	return m.get(where, usr.ID)
}

func (m Model) UpdateUserName(tkn, newUserName string) error {
	t, err := m.token.Validate(tkn)
	if err != nil {
		return err
	}
	if newUserName == "" {
		return errors.NewClient("the userName provided was invlaid")
	}
	q := `SELECT COUNT(id) FROM userNames WHERE userID=$1`
	var count int
	if m.db.QueryRow(q, t.UserID()).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has usernae: %s", err)
	}
	if (count == 0) {
		q = `INSERT INTO userNames (userID, userName, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
		rslt, err := m.db.Exec(q, t.UserID(), newUserName)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE userNames
		 	SET userName=$1, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$2`
	rslt, err := m.db.Exec(q, newUserName, t.UserID())
	return checkRowsAffected(rslt, err, 1)
}

func (m Model) UpdateAppUserID(tkn string, new *authms.OAuth) error {
	t, err := m.token.Validate(tkn)
	if err != nil {
		return err
	}
	if err = validateOAuth(new); err != nil {
		return err
	}
	q := `SELECT COUNT(id) FROM appUserIDs WHERE userID=$1 AND appName=$2`
	var count int
	if m.db.QueryRow(q, t.UserID(), new.AppName).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has email: %s", err)
	}
	if (count == 0) {
		q = `INSERT INTO appUserIDs (userID, appUserID, appName, createDate)
	 		VALUES ($1, $2, $3, CURRENT_TIMESTAMP())`
		rslt, err := m.db.Exec(q, t.UserID(), new.AppUserID, new.AppName)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE appUserIDs
		 	SET appUserID=$1, validated=$2, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$3 AND appName=$4`
	rslt, err := m.db.Exec(q, new.AppUserID, false, t.UserID(), new.AppName)
	return checkRowsAffected(rslt, err, 1)
}

func (m Model) UpdateEmail(tkn, newEmail string) error {
	t, err := m.token.Validate(tkn)
	if err != nil {
		return err
	}
	if newEmail == "" {
		return errors.NewClient("the email provided was invlaid")
	}
	q := `SELECT COUNT(id) FROM emails WHERE userID=$1`
	var count int
	if m.db.QueryRow(q, t.UserID()).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has email: %s", err)
	}
	if (count == 0) {
		q = `INSERT INTO emails (userID, email, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
		rslt, err := m.db.Exec(q, t.UserID(), newEmail)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE emails
		 	SET email=$1, validated=$2, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$3`
	rslt, err := m.db.Exec(q, newEmail, false, t.UserID())
	return checkRowsAffected(rslt, err, 1)
}

func (m Model) UpdatePhone(tkn, newPhone string) error {
	t, err := m.token.Validate(tkn)
	if err != nil {
		return err
	}
	if newPhone == "" {
		return errors.NewClient("the phone provided was invlaid")
	}
	q := `SELECT COUNT(id) FROM phones WHERE userID=$1`
	var count int
	if m.db.QueryRow(q, t.UserID()).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has phone: %s", err)
	}
	if (count == 0) {
		q = `INSERT INTO phones (userID, phone, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
		rslt, err := m.db.Exec(q, t.UserID(), newPhone)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE phones
		 	SET phone=$1, validated=$2, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$3`
	rslt, err := m.db.Exec(q, newPhone, false, t.UserID())
	return checkRowsAffected(rslt, err, 1)
}

func (m Model) UpdatePassword(userID int64, oldPass, newPassword string) error {
	q := `SELECT password FROM users WHERE id=$1`
	var currPass []byte
	err := m.db.QueryRow(q, userID).Scan(&currPass)
	if err != nil {
		if err != sql.ErrNoRows {
			return ErrorPasswordMismatch
		}
		return err
	}
	if ! m.hasher.CompareHash(newPassword, currPass) {
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

func checkRowsAffected(rslt sql.Result, err error, expAffected int64) error {
	if err != nil {
		return err
	}
	c, err := rslt.RowsAffected()
	if err != nil {
		return err
	}
	if c != expAffected {
		return errors.Newf("expected %d affected rows but got %d",
			expAffected, c)
	}
	return nil
}

func (m Model) get(where string, whereArgs... interface{}) (*authms.User, error) {
	usr := &authms.User{
		Email: &authms.Value{},
		Phone: &authms.Value{},
		OAuth: &authms.OAuth{},
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
	query = `SELECT appUserID, appName FROM appUserIDs WHERE userID=$1`
	rslt, err := m.db.Query(query, usr.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return usr, nil
		}
		return usr, err
	}
	apps := make([]*authms.OAuth, 0)
	for rslt.Next() {
		app := new(authms.OAuth)
		if err := rslt.Scan(&app.AppUserID, &app.AppName); err != nil {
			return usr, err
		}
		apps = append(apps, app)
	}
	if err = rslt.Close(); err != nil {
		return usr, err
	}
	usr.OAuth = apps[0]
	return usr, nil
}

func (m Model) validatePassword(id int64, password string) error {
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

func (m *Model) getPasswordHash(u *authms.User) ([]byte, error) {
	passStr := u.Password
	if passStr == "" && !havePasswordComboAuth(u) {
		passB, err := m.gen.SecureRandomString(36)
		if err != nil {
			return nil, err
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

func havePasswordComboAuth(u *authms.User) bool {
	return HasValue(u.Phone) || HasValue(u.Email) || u.UserName != ""
}

func validateUser(u *authms.User) error {
	if u == nil {
		return errors.NewClient("user was not provided")
	}
	hasPhone := HasValue(u.Phone)
	hasMail := HasValue(u.Email)
	if u.UserName == "" && !hasPhone && !hasMail && u.OAuth == nil {
		return errors.NewClient("A user must have at least one" +
			" identifier (UserName, Phone, Email, OAuthApp")
	}
	if havePasswordComboAuth(u) && u.Password == "" {
		return ErrorEmptyPassword
	}
	if u.OAuth != nil {
		if err := validateOAuth(u.OAuth); err != nil {
			return err
		}
	}
	return nil
}

func validateOAuth(oa *authms.OAuth) error {
	if oa == nil {
		return errors.NewClient("OAuth was not provided")
	}
	if oa.AppName == "" {
		return errors.NewClient("AppName was not provided")
	}
	if oa.AppToken == "" {
		return errors.NewClient("AppToken was not provided")
	}
	if oa.AppUserID == "" {
		return errors.NewClient("AppUserID was not provided")
	}
	return nil
}

func HasValue(v *authms.Value) bool {
	return v != nil && v.Value != ""
}
