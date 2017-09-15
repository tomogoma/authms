package dbhelper

import (
	"database/sql"

	"sync"

	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/go-commons/errors"
)

type DBHelper struct {
	hasher    Hasher
	gen       PasswordGenerator
	gcRunning bool
	errors.NotFoundErrCheck
	dsnF cockroach.DSNFormatter

	isDBInitMutex sync.Mutex
	isDBInit      bool
	db            *sql.DB
}

func New(dsnF cockroach.DSNFormatter, pg PasswordGenerator, h Hasher) (*DBHelper, error) {
	if h == nil {
		return nil, errors.New("HashFunc cannot be nil")
	}
	if pg == nil {
		return nil, errors.New("HashFunc cannot be nil")
	}
	if dsnF == nil {
		return nil, errors.New("DSNFormatter was nil")
	}
	return &DBHelper{dsnF: dsnF, gen: pg, hasher: h, isDBInitMutex: sync.Mutex{}}, nil
}

func (h *DBHelper) initDBConnIfNotInitted() error {
	var err error
	h.db, err = cockroach.TryConnect(h.dsnF.FormatDSN(), h.db)
	if err != nil {
		return errors.Newf("Failed to connect to db: %v", err)
	}
	return h.instantiate()
}

func (h *DBHelper) instantiate() error {
	h.isDBInitMutex.Lock()
	defer h.isDBInitMutex.Unlock()
	if h.isDBInit {
		return nil
	}
	if err := cockroach.InstantiateDB(h.db, h.dsnF.DBName(), AllTableDescs...); err != nil {
		return errors.Newf("error instantiating db: %v", err)
	}
	h.isDBInit = true
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
