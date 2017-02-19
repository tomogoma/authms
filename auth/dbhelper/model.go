package dbhelper

import (
	"github.com/tomogoma/go-commons/errors"
	"github.com/tomogoma/go-commons/database/cockroach"
	"database/sql"
	"github.com/tomogoma/go-commons/auth/token"
)

type DBHelper struct {
	db          *sql.DB
	hasher      Hasher
	gen         PasswordGenerator
	token       TokenValidator
	tokenSaveCh chan *token.Token
	tokenDelCh  chan string
	gcRunning   bool
}

var ErrorNilHashFunc = errors.New("HashFunc cannot be nil")
var ErrorNilPasswordGenerator = errors.New("password generator was nil")
var ErrorNilTokenValidator = errors.New("token validator was nil")

func New(dsnF cockroach.DSNFormatter, pg PasswordGenerator, h Hasher, tv TokenValidator) (*DBHelper, error) {
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
	if err := cockroach.InstantiateDB(db, dsnF.DBName(), users, usernames,
		emails, phones, appUserIDs, history); err != nil {
		return nil, errors.Newf("error instantiating db: %s", err)
	}
	iCh := make(chan *token.Token)
	dCh := make(chan string)
	return &DBHelper{db: db, gen: pg, hasher: h, token: tv, tokenSaveCh: iCh, tokenDelCh: dCh}, nil
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
