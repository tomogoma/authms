package dbhelper

import (
	"github.com/tomogoma/go-commons/errors"
	"github.com/tomogoma/go-commons/database/cockroach"
	"database/sql"
)

type DBHelper struct {
	db        *sql.DB
	hasher    Hasher
	gen       PasswordGenerator
	gcRunning bool
	errors.NotFoundErrCheck
}

var ErrorNilHashFunc = errors.New("HashFunc cannot be nil")
var ErrorNilPasswordGenerator = errors.New("password generator was nil")

func New(dsnF cockroach.DSNFormatter, pg PasswordGenerator, h Hasher) (*DBHelper, error) {
	if h == nil {
		return nil, ErrorNilHashFunc
	}
	if pg == nil {
		return nil, ErrorNilPasswordGenerator
	}
	db, err := cockroach.DBConn(dsnF)
	if err != nil {
		return nil, errors.Newf("error connecting to db: %s", err)
	}
	if err := cockroach.InstantiateDB(db, dsnF.DBName(), users, usernames,
		emails, phones, appUserIDs, history); err != nil {
		return nil, errors.Newf("error instantiating db: %s", err)
	}
	return &DBHelper{db: db, gen: pg, hasher: h}, nil
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
