package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/go-commons/errors"
)

// Roach is a cockroach db store.
// Use NewRoach() to instantiate.
type Roach struct {
	errors.NotFoundErrCheck
	dsn              string
	dbName           string
	db               *sql.DB
	compatibilityErr error

	isDBInitMutex sync.Mutex
	isDBInit      bool
}

const (
	keyDBVersion = "db.version"
	keySMTPConf  = "conf.smtp"
)

var errorNilTx = errors.Newf("sql Tx was nil")

// NewRoach creates an instance of *Roach. A db connection is only established
// when InitDBIfNot() or one of the Execute/Query methods is called.
func NewRoach(opts ...Option) *Roach {
	r := &Roach{
		isDBInit:      false,
		isDBInitMutex: sync.Mutex{},
		dbName:        config.CanonicalName,
	}
	for _, f := range opts {
		f(r)
	}
	return r
}

// InitDBIfNot connects to and sets up the DB; creating it and tables if necessary.
func (r *Roach) InitDBIfNot() error {
	var err error
	r.db, err = cockroach.TryConnect(r.dsn, r.db)
	if err != nil {
		return errors.Newf("connect to db: %v", err)
	}
	return r.instantiate()
}

// UpsertSMTPConfig upserts SMTP config values into the db.
func (r *Roach) UpsertSMTPConfig(conf interface{}) error {
	return r.upsertConf(keySMTPConf, conf)
}

// GetSMTPConfig fetches SMTP config values from the db and unmarshals them
// into conf. this method fails if conf is nil or not a pointer.
func (r *Roach) GetSMTPConfig(conf interface{}) error {
	return r.getConf(keySMTPConf, conf)
}

// ExecuteTx prepares a transaction (with retries) for execution in fn.
// It commits the changes if fn returns nil, otherwise changes are rolled back.
func (r *Roach) ExecuteTx(fn func(*sql.Tx) error) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
	return crdb.ExecuteTx(context.Background(), r.db, nil, fn)
}

// InsertGroup inserts into the database returning calculated values.
func (r *Roach) InsertGroup(name string, acl int) (*model.Group, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	grp := model.Group{Name: name, AccessLevel: acl}
	insCols := ColDesc(ColName, ColAccessLevel, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblGroups + ` (` + insCols + `)
		VALUES ($1,$2,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := r.db.QueryRow(q, name, acl).Scan(&grp.ID, &grp.CreateDate, &grp.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &grp, nil
}
func (r *Roach) Group(string) (*model.Group, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) GroupByName(string) (*model.Group, error) {
	return nil, errors.NewNotImplemented()
}

// InsertUserType inserts into the database returning calculated values.
func (r *Roach) InsertUserType(name string) (*model.UserType, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	ut := model.UserType{Name: name}
	insCols := ColDesc(ColName, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblUserTypes + ` (` + insCols + `)
		VALUES ($1,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := r.db.QueryRow(q, name).Scan(&ut.ID, &ut.CreateDate, &ut.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &ut, nil
}
func (r *Roach) UserTypeByName(string) (*model.UserType, error) {
	return nil, errors.NewNotImplemented()
}

// InsertUserType inserts into the database returning calculated values.
func (r *Roach) InsertUserAtomic(tx *sql.Tx, typeID string, password []byte) (*model.User, error) {
	if tx == nil {
		return nil, errorNilTx
	}
	u := model.User{Type: model.UserType{ID: typeID}}
	insCols := ColDesc(ColTypeID, ColPassword, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblUsers + ` (` + insCols + `)
		VALUES ($1,$2,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, typeID, password).Scan(&u.ID, &u.CreateDate, &u.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *Roach) UpdatePassword(userID string, password []byte) error {
	return errors.NewNotImplemented()
}
func (r *Roach) UpdatePasswordAtomic(tx *sql.Tx, userID string, password []byte) error {
	return errors.NewNotImplemented()
}
func (r *Roach) User(id string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByDeviceID(devID string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByUsername(username string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByPhone(phone string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByEmail(email string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByFacebook(facebookID string) (*model.User, error) {
	return nil, errors.NewNotImplemented()
}

func (r *Roach) AddUserToGroupAtomic(tx *sql.Tx, userID, groupID string) error {
	return errors.NewNotImplemented()
}

func (r *Roach) InsertUserDeviceAtomic(tx *sql.Tx, userID, devID string) (*model.Device, error) {
	if tx == nil {
		return nil, errorNilTx
	}
	dev := model.Device{UserID: userID, DeviceID: devID}
	insCols := ColDesc(ColUserID, ColDevID, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblDeviceIDs + ` (` + insCols + `)
		VALUES ($1,$2,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, devID).Scan(&dev.ID, &dev.CreateDate, &dev.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

func (r *Roach) InsertUserName(userID, username string) (*model.Username, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) InsertUserNameAtomic(tx *sql.Tx, userID, username string) (*model.Username, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) UpdateUsername(userID, username string) (*model.Username, error) {
	return nil, errors.NewNotImplemented()
}

func (r *Roach) InsertUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) UpdateUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) UpdateUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}

func (r *Roach) InsertPhoneToken(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) InsertPhoneTokenAtomic(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) PhoneTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	return nil, errors.NewNotImplemented()
}

func (r *Roach) InsertUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) UpdateUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) UpdateUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}

func (r *Roach) InsertEmailToken(userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) InsertEmailTokenAtomic(tx *sql.Tx, userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) EmailTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	return nil, errors.NewNotImplemented()
}

func (r *Roach) InsertUserFbIDAtomic(tx *sql.Tx, userID, fbID string, verified bool) (*model.Facebook, error) {
	if tx == nil {
		return nil, errorNilTx
	}
	fb := model.Facebook{UserID: userID, FacebookID: fbID, Verified: verified}
	insCols := ColDesc(ColUserID, ColFacebookID, ColVerified, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblUsers + ` (` + insCols + `)
		VALUES ($1,$2,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, fbID, verified).Scan(&fb.ID, &fb.CreateDate, &fb.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &fb, nil
}

// ColDesc returns a string containing cols in the given order separated by ",".
func ColDesc(cols ...string) string {
	desc := ""
	for _, col := range cols {
		desc = desc + col + ","
	}
	return strings.TrimSuffix(desc, ",")
}

func (r *Roach) instantiate() error {
	r.isDBInitMutex.Lock()
	defer r.isDBInitMutex.Unlock()
	if r.compatibilityErr != nil {
		return r.compatibilityErr
	}
	if r.isDBInit {
		return nil
	}
	if err := cockroach.InstantiateDB(r.db, r.dbName, AllTableDescs...); err != nil {
		return errors.Newf("instantiating db: %v", err)
	}
	if err := r.validateRunningVersion(); err != nil {
		if !r.IsNotFoundError(err) {
			return fmt.Errorf("check db version: %v", err)
		}
		if err := r.setRunningVersionCurrent(); err != nil {
			return errors.Newf("set db version: %v", err)
		}
	}
	r.isDBInit = true
	return nil
}

func (r *Roach) validateRunningVersion() error {
	var runningVersion int
	q := `SELECT ` + ColValue + ` FROM ` + TblConfigurations + ` WHERE ` + ColKey + `=$1`
	var confB []byte
	if err := r.db.QueryRow(q, keyDBVersion).Scan(&confB); err != nil {
		if err == sql.ErrNoRows {
			return errors.NewNotFoundf("config not found")
		}
		return errors.Newf("get conf: %v", err)
	}
	if err := json.Unmarshal(confB, &runningVersion); err != nil {
		return errors.Newf("Unmarshalling config: %v", err)
	}
	if runningVersion != Version {
		r.compatibilityErr = errors.Newf("db incompatible: need db"+
			" version '%d', found '%d'", Version, runningVersion)
		return r.compatibilityErr
	}
	return nil
}

func (r *Roach) setRunningVersionCurrent() error {
	valB, err := json.Marshal(Version)
	if err != nil {
		return errors.Newf("marshal conf: %v", err)
	}
	cols := ColDesc(ColKey, ColValue, ColUpdateDate)
	q := `UPSERT INTO ` + TblConfigurations + ` (` + cols + `) VALUES ($1, $2, CURRENT_TIMESTAMP)`
	res, err := r.db.Exec(q, keyDBVersion, valB)
	return checkRowsAffected(res, err, 1)
}

func (r *Roach) upsertConf(key string, conf interface{}) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
	valB, err := json.Marshal(conf)
	if err != nil {
		return errors.Newf("marshal conf: %v", err)
	}
	cols := ColDesc(ColKey, ColValue, ColUpdateDate)
	q := `UPSERT INTO ` + TblConfigurations + ` (` + cols + `) VALUES ($1, $2, CURRENT_TIMESTAMP)`
	res, err := r.db.Exec(q, key, valB)
	return checkRowsAffected(res, err, 1)
}

func (r *Roach) getConf(key string, conf interface{}) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
	q := `SELECT ` + ColValue + ` FROM ` + TblConfigurations + ` WHERE ` + ColKey + `=$1`
	var confB []byte
	if err := r.db.QueryRow(q, key).Scan(&confB); err != nil {
		if err == sql.ErrNoRows {
			return errors.NewNotFoundf("config not found")
		}
		return err
	}
	if err := json.Unmarshal(confB, conf); err != nil {
		return errors.Newf("Unmarshalling config: %v", err)
	}
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
