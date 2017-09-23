package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/lib/pq"
	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/go-commons/errors"
	"golang.org/x/crypto/bcrypt"
)

type PasswordGenerator interface {
	SecureRandomBytes(length int) ([]byte, error)
}

type Roach struct {
	gen       PasswordGenerator
	gcRunning bool
	errors.NotFoundErrCheck
	dsnF cockroach.DSNFormatter

	isDBInitMutex sync.Mutex
	isDBInit      bool
	db            *sql.DB
}

var ErrorPasswordMismatch = errors.NewAuth("username/password combo mismatch")
var ErrorModelCorruptedOnEmptyPassword = errors.New("The model contained an empty password value and is probably corrupt")

func NewRoach(dsnF cockroach.DSNFormatter, pg PasswordGenerator) (*Roach, error) {
	if pg == nil {
		return nil, errors.New("HashFunc cannot be nil")
	}
	if dsnF == nil {
		return nil, errors.New("DSNFormatter was nil")
	}
	return &Roach{dsnF: dsnF, gen: pg, isDBInitMutex: sync.Mutex{}}, nil
}

func (r *Roach) InitDBConnIfNotInitted() error {
	var err error
	r.db, err = cockroach.TryConnect(r.dsnF.FormatDSN(), r.db)
	if err != nil {
		return errors.Newf("Failed to connect to db: %v", err)
	}
	return r.instantiate()
}

func (r *Roach) SaveHistory(h authms.History) (*authms.History, error) {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return nil, err
	}
	q := `
	INSERT INTO history (userID, accessMethod, successful, devID, ipAddress, date)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP())
		 RETURNING id
	`
	err := r.db.QueryRow(q, h.UserID, h.AccessType, h.SuccessStatus,
		h.DevID, h.IpAddress).Scan(&h.ID)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *Roach) GetHistory(userID int64, offset, count int, acMs ...string) ([]*authms.History, error) {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return nil, err
	}
	acMFilter := ""
	for i, acM := range acMs {
		if i == 0 {
			acMFilter = fmt.Sprintf("AND (accessMethod = '%s'", acM)
			continue
		}
		acMFilter = fmt.Sprintf("%s OR accessMethod = '%s'", acMFilter, acM)
	}
	if acMFilter != "" {
		acMFilter += ")"
	}
	q := fmt.Sprintf(`
		SELECT id, accessMethod, successful, userID, date, devID, ipAddress
		FROM history
		WHERE userID = $1 %s
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`, acMFilter)
	rows, err := r.db.Query(q, userID, count, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hists := make([]*authms.History, 0)
	for rows.Next() {
		var devID, ipAddr sql.NullString
		d := &authms.History{}
		err = rows.Scan(&d.ID, &d.AccessType, &d.SuccessStatus, &d.UserID,
			&d.Date, &devID, &ipAddr)
		d.DevID = devID.String
		d.IpAddress = ipAddr.String
		if err != nil {
			return nil, errors.Newf("scanning row: %v", err)
		}
		hists = append(hists, d)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterating resultset: %v", err)
	}
	if len(hists) == 0 {
		return nil, errors.NewNotFound("no history item found")
	}
	return hists, nil
}

func (r *Roach) SaveUser(u authms.User) (*authms.User, error) {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return nil, err
	}
	passHB, err := r.getPasswordHash(u)
	if err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err = crdb.ExecuteTx(ctx, r.db, nil, func(tx *sql.Tx) error {
		userqStr := `INSERT INTO users (password, createDate)
	 		VALUES ($1, CURRENT_TIMESTAMP()) RETURNING id`
		err := tx.QueryRow(userqStr, passHB).Scan(&u.ID)
		if err != nil {
			return err
		}
		if hasValue(u.Email) {
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
		if hasValue(u.Phone) {
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
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// returns userID for existence of identifier in the db for any of the user types
// otherwise returns an error.
// Roach#IsNotFoundError(err) returns true if on the returned error if the user
// does not exist.
func (r *Roach) UserExists(u authms.User) (int64, error) {
	userID := int64(-1)
	var err error
	if err = r.InitDBConnIfNotInitted(); err != nil {
		return userID, err
	}
	if u.UserName != "" {
		q := `SELECT userID FROM userNames WHERE userName=$1`
		err = r.db.QueryRow(q, u.UserName).Scan(&userID)
		if err == nil || err != sql.ErrNoRows {
			return userID, err
		}
	}
	if hasValue(u.Email) {
		q := `SELECT userID FROM emails WHERE email=$1`
		err = r.db.QueryRow(q, u.Email.Value).Scan(&userID)
		if err == nil || err != sql.ErrNoRows {
			return userID, err
		}
	}
	if hasValue(u.Phone) {
		q := `SELECT userID FROM phones WHERE phone=$1`
		err = r.db.QueryRow(q, u.Phone.Value).Scan(&userID)
		if err == nil || err != sql.ErrNoRows {
			return userID, err
		}
	}
	for _, oAuth := range u.OAuths {
		q := `SELECT userID FROM appUserIDs WHERE appName=$1 AND appUserID=$2`
		err = r.db.QueryRow(q, oAuth.AppName, oAuth.AppUserID).Scan(&userID)
		if err == nil || err != sql.ErrNoRows {
			return userID, err
		}
	}
	if err == sql.ErrNoRows {
		return userID, errors.NewNotFound("User does not exist")
	}
	return userID, nil
}

func (r *Roach) GetByUserName(uName, pass string) (*authms.User, error) {
	where := "usernames.userName = $1"
	usr, err := r.get(where, uName)
	return r.validateFetchedUser(usr, err, pass)
}

func (r *Roach) GetByPhone(phone, pass string) (*authms.User, error) {
	where := `phones.phone = $1`
	usr, err := r.get(where, phone)
	return r.validateFetchedUser(usr, err, pass)
}

func (r *Roach) GetByEmail(email, pass string) (*authms.User, error) {
	where := `emails.email = $1`
	usr, err := r.get(where, email)
	return r.validateFetchedUser(usr, err, pass)
}

func (r *Roach) GetByAppUserID(appName, appUserID string) (*authms.User, error) {
	usr := &authms.User{}
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return usr, err
	}
	query := `SELECT userID FROM appUserIDs WHERE appName = $1 AND appUserID = $2`
	err := r.db.QueryRow(query, appName, appUserID).Scan(&usr.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			return usr, err
		}
		return usr, ErrorPasswordMismatch
	}
	where := "users.id=$1"
	return r.get(where, usr.ID)
}

func (r *Roach) UpdateUserName(userID int64, newUserName string) error {
	if newUserName == "" {
		return errors.New("the userName provided was invlaid")
	}
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return err
	}
	q := `SELECT COUNT(id) FROM userNames WHERE userID=$1`
	var count int
	if err := r.db.QueryRow(q, userID).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has usernae: %s", err)
	}
	if count == 0 {
		q = `INSERT INTO userNames (userID, userName, createDate)
		 		VALUES ($1, $2, CURRENT_TIMESTAMP())`
		rslt, err := r.db.Exec(q, userID, newUserName)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE userNames
		 	SET userName=$1, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$2`
	rslt, err := r.db.Exec(q, newUserName, userID)
	return checkRowsAffected(rslt, err, 1)
}

func (r *Roach) UpdateAppUserID(userID int64, new authms.OAuth) error {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return err
	}
	q := `SELECT COUNT(id) FROM appUserIDs WHERE userID=$1 AND appName=$2`
	var count int
	if err := r.db.QueryRow(q, userID, new.AppName).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has email: %s", err)
	}
	if count == 0 {
		q = `INSERT INTO appUserIDs (userID, appUserID, appName, validated, createDate)
	 		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP())`
		rslt, err := r.db.Exec(q, userID, new.AppUserID, new.AppName, new.Verified)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE appUserIDs
		 	SET appUserID=$1, validated=$2, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$3 AND appName=$4`
	rslt, err := r.db.Exec(q, new.AppUserID, new.Verified, userID, new.AppName)
	return checkRowsAffected(rslt, err, 1)
}

func (r *Roach) UpdateEmail(userID int64, newEmail authms.Value) error {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return err
	}
	q := `SELECT COUNT(id) FROM emails WHERE userID=$1`
	var count int
	if err := r.db.QueryRow(q, userID).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has email: %s", err)
	}
	if count == 0 {
		q = `INSERT INTO emails (userID, email, validated, createDate)
		 		VALUES ($1, $2, $3, CURRENT_TIMESTAMP())`
		rslt, err := r.db.Exec(q, userID, newEmail.Value, newEmail.Verified)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE emails
		 	SET email=$1, validated=$2, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$3`
	rslt, err := r.db.Exec(q, newEmail.Value, newEmail.Verified, userID)
	return checkRowsAffected(rslt, err, 1)
}

func (r *Roach) UpdatePhone(userID int64, newPhone authms.Value) error {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return err
	}
	q := `SELECT COUNT(id) FROM phones WHERE userID=$1`
	var count int
	if err := r.db.QueryRow(q, userID).Scan(&count); err != nil {
		return fmt.Errorf("error checking if user has phone: %s", err)
	}
	if count == 0 {
		q = `INSERT INTO phones (userID, phone, validated, createDate)
		 		VALUES ($1, $2, $3, CURRENT_TIMESTAMP())`
		rslt, err := r.db.Exec(q, userID, newPhone.Value, newPhone.Verified)
		return checkRowsAffected(rslt, err, 1)
	}
	q = `UPDATE phones
		 	SET phone=$1, validated=$2, updateDate=CURRENT_TIMESTAMP()
		 	WHERE userID=$3`
	rslt, err := r.db.Exec(q, newPhone.Value, newPhone.Verified, userID)
	return checkRowsAffected(rslt, err, 1)
}

func (r *Roach) UpsertLoginVerification(lv model.LoginVerification) (model.LoginVerification, error) {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return lv, err
	}
	if lv.ID == "" {
		lv.ID = genID()
	}
	q := `
	UPSERT INTO authVerifications (id, type, subjectValue, userID,
			codeHash, isUsed, issueDate, expiryDate)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	rslt, err := r.db.Exec(q, lv.ID, lv.Type, lv.SubjectValue, lv.UserID, lv.CodeHash,
		lv.IsUsed, lv.Issue, lv.Expiry)
	return lv, checkRowsAffected(rslt, err, 1)
}

func (r *Roach) GetLoginVerifications(vType string, userID, offset, count int64) ([]model.LoginVerification, error) {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return nil, err
	}
	q := `
	SELECT id, type, subjectValue, userID, codeHash, isUsed, issueDate, expiryDate
		FROM authVerifications
		WHERE type=$1 AND userID=$2
		LIMIT $3 OFFSET $4
		ORDER BY isused ASC, expirydate DESC
	`
	rows, err := r.db.Query(q, vType, userID, count, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lvs := make([]model.LoginVerification, 0)
	for rows.Next() {
		lv := model.LoginVerification{}
		err = rows.Scan(&lv.ID, &lv.Type, &lv.SubjectValue, &lv.UserID, &lv.CodeHash,
			&lv.IsUsed, &lv.Issue, &lv.Expiry)
		if err != nil {
			return nil, errors.Newf("scan row: %v", err)
		}
		lvs = append(lvs, lv)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Newf("iterate resultset: %v", err)
	}
	if len(lvs) == 0 {
		return nil, errors.NewNotFound("no verification found")
	}
	return lvs, nil
}

func (r *Roach) UpdatePassword(userID int64, oldPass, newPassword string) error {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return err
	}
	q := `SELECT password FROM users WHERE id=$1`
	var actPassHB []byte
	err := r.db.QueryRow(q, userID).Scan(&actPassHB)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("not found")
			return ErrorPasswordMismatch
		}
		return err
	}
	if err := bcrypt.CompareHashAndPassword(actPassHB, []byte(oldPass)); err != nil {
		return ErrorPasswordMismatch
	}
	passHB, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	q = `UPDATE users
			SET password=$1, updateDate=CURRENT_TIMESTAMP()
	 		WHERE id=$2`
	rslt, err := r.db.Exec(q, passHB, userID)
	return checkRowsAffected(rslt, err, 1)
}

func (r *Roach) IsDuplicateError(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505"
}

func (r *Roach) validateFetchedUser(usr *authms.User, getErr error, pass string) (
	*authms.User, error) {
	if getErr != nil {
		return usr, getErr
	}
	if getErr = r.validatePassword(usr.ID, pass); getErr != nil {
		return usr, getErr
	}
	return usr, nil
}

func (r *Roach) get(where string, whereArgs ...interface{}) (*authms.User, error) {
	usr := &authms.User{
		Email:  &authms.Value{},
		Phone:  &authms.Value{},
		OAuths: make(map[string]*authms.OAuth),
	}
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return usr, err
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
	err := r.db.QueryRow(query, whereArgs...).Scan(&usr.ID, &dbUserName,
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
	rslt, err := r.db.Query(query, usr.ID)
	if err != nil {
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

func (r *Roach) validatePassword(id int64, password string) error {
	if err := r.InitDBConnIfNotInitted(); err != nil {
		return err
	}
	userQ := `SELECT password FROM users WHERE id = $1`
	var dbPassword []byte
	err := r.db.QueryRow(userQ, id).Scan(&dbPassword)
	if err != nil {
		return err
	}
	if len(dbPassword) == 0 {
		return ErrorModelCorruptedOnEmptyPassword
	}
	if err := bcrypt.CompareHashAndPassword(dbPassword, []byte(password)); err != nil {
		return ErrorPasswordMismatch
	}
	return err
}

func (r *Roach) getPasswordHash(u authms.User) ([]byte, error) {
	passStr := u.Password
	if passStr == "" {
		passB, err := r.gen.SecureRandomBytes(36)
		if err != nil {
			return nil, errors.Newf("error generating password: %v",
				err)
		}
		passStr = string(passB)
	}
	return bcrypt.GenerateFromPassword([]byte(passStr), bcrypt.DefaultCost)
}

func (r *Roach) instantiate() error {
	r.isDBInitMutex.Lock()
	defer r.isDBInitMutex.Unlock()
	if r.isDBInit {
		return nil
	}
	if err := cockroach.InstantiateDB(r.db, r.dsnF.DBName(), AllTableDescs...); err != nil {
		return errors.Newf("error instantiating db: %v", err)
	}
	r.isDBInit = true
	return nil
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

func validateHistory(h *authms.History) error {
	if h.UserID < 1 {
		return errors.New("userID was invalid")
	}
	if h.AccessType == "" {
		return errors.New("access type was empty")
	}
	return nil
}

func hasValue(v *authms.Value) bool {
	return v != nil && v.Value != ""
}

func genID() string {
	return uuid.New()
}
