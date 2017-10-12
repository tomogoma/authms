package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/tomogoma/authms/config"
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

type inserter interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
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

// ExecuteTx prepares a transaction (with retries) for execution in fn.
// It commits the changes if fn returns nil, otherwise changes are rolled back.
func (r *Roach) ExecuteTx(fn func(*sql.Tx) error) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
	return crdb.ExecuteTx(context.Background(), r.db, nil, fn)
}

// ColDesc returns a string containing cols in the given order separated by ",".
func ColDesc(cols ...string) string {
	desc := ""
	for _, col := range cols {
		desc = desc + col + ","
	}
	return strings.TrimSuffix(desc, ",")
}

func colDescTbl(tbl string, cols ...string) string {
	desc := ""
	for _, col := range cols {
		desc = desc + tbl + "." + col + ","
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
	updCols := ColDesc(ColValue, ColUpdateDate)
	q := `
		INSERT INTO ` + TblConfigurations + ` (` + cols + `)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT (` + ColKey + `)
			DO UPDATE SET (` + updCols + `) = ($2, CURRENT_TIMESTAMP)`
	res, err := r.db.Exec(q, keyDBVersion, valB)
	return checkRowsAffected(res, err, 1)
}

func checkRowsAffected(r sql.Result, err error, expAffected int64) error {
	if err != nil {
		return err
	}
	c, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if c != expAffected {
		return errors.Newf("expected %d affected rows but got %d",
			expAffected, c)
	}
	return nil
}
