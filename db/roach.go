package db

import (
	"database/sql"
	"sync"

	"strconv"

	"fmt"

	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/go-commons/errors"
)

type Roach struct {
	errors.NotFoundErrCheck
	dsn              string
	dbName           string
	db               *sql.DB
	compatibilityErr error

	isDBInitMutex sync.Mutex
	isDBInit      bool
}

type Option func(*Roach)

const (
	KeyConfigDBVersion = "db.version"
)

func WithDSN(dsn string) Option {
	return func(r *Roach) {
		r.dsn = dsn
	}
}

func WithDBName(db string) Option {
	return func(r *Roach) {
		r.dbName = db
	}
}

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

func (r *Roach) InitDBConnIfNotInitted() error {
	var err error
	r.db, err = cockroach.TryConnect(r.dsn, r.db)
	if err != nil {
		return errors.Newf("connect to db: %v", err)
	}
	return r.instantiate()
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
		if err := r.upsertDBVersion(); err != nil {
			return errors.Newf("set db version: %v", err)
		}
	}
	r.isDBInit = true
	return nil
}

func (r *Roach) upsertDBVersion() error {
	q := `
		UPSERT INTO ` + TblConfigurations + ` (` + ColKey + `, ` + ColValue + `)
			VALUES ('` + KeyConfigDBVersion + `', '` + strconv.Itoa(Version) + `')`
	res, err := r.db.Exec(q)
	return checkRowsAffected(res, err, 1)
}

func (r *Roach) validateRunningVersion() error {
	runningVersionStr := ""
	q := `
	SELECT ` + ColValue + `
		FROM ` + TblConfigurations + `
		WHERE ` + ColKey + `=` + KeyConfigDBVersion
	err := r.db.QueryRow(q).Scan(&runningVersionStr)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return errors.NewNotFound("version not set")
	}
	runningVersion, err := strconv.Atoi(runningVersionStr)
	if err != nil || runningVersion != Version {
		r.compatibilityErr = errors.Newf("db incompatible: need db"+
			" version '%d', found '%s'", Version, runningVersionStr)
		return r.compatibilityErr
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

func hasValue(v *authms.Value) bool {
	return v != nil && v.Value != ""
}

func genID() string {
	return uuid.New()
}
