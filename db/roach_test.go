package db_test

import (
	"database/sql"
	"strconv"
	"testing"

	"github.com/tomogoma/authms/db"
	testingH "github.com/tomogoma/authms/testing"
	"github.com/tomogoma/crdb"
	errors "github.com/tomogoma/go-typed-errors"
)

var isInit bool

func setup(t *testing.T) crdb.Config {
	conf := testingH.ReadConfig(t)
	conf.Database.DBName = conf.Database.DBName + "_test"
	if !isInit {
		rdb := getDB(t, conf.Database)
		defer rdb.Close()
		err := dropAllTables(rdb, conf.Database.DBName)
		if err != nil {
			t.Fatalf("Error setting up: drop prev test tables: %v", err)
		}
		isInit = true
	}
	return conf.Database
}

func tearDown(t *testing.T, conf crdb.Config) {
	rdb := getDB(t, conf)
	defer rdb.Close()
	if err := delAllTables(rdb, conf.DBName); err != nil {
		t.Fatalf("Error tearing down: delete all tables: %v", err)
	}
}

func dbCreated(rdb *sql.DB, dbName string) (bool, error) {
	rslt, err := rdb.Query("SHOW databases")
	if err != nil {
		return false, errors.Newf("list databases: %v", err)
	}
	defer rslt.Close()
	for rslt.Next() {
		var name string
		if err := rslt.Scan(&name); err != nil {
			return false, errors.Newf("dbName from resultset: %v", err)
		}
		if dbName == name {
			return true, nil
		}
	}
	if err := rslt.Err(); err != nil {
		return false, errors.Newf("iterating resultset: %v", err)
	}
	return false, nil
}

func delAllTables(rdb *sql.DB, dbName string) error {
	haveDB, err := dbCreated(rdb, dbName)
	if err != nil {
		return errors.Newf("check test db created: %v", err)
	}
	if !haveDB {
		return nil
	}
	for i := len(db.AllTableNames) - 1; i >= 0; i-- {
		_, err := rdb.Exec("DELETE FROM " + db.AllTableNames[i])
		if err != nil {
			return errors.Newf("delete %s: %v", db.AllTableNames[i], err)
		}
	}
	return nil
}

func dropAllTables(rdb *sql.DB, dbName string) error {
	haveDB, err := dbCreated(rdb, dbName)
	if err != nil {
		return errors.Newf("check test db created: %v", err)
	}
	if !haveDB {
		return nil
	}
	for i := len(db.AllTableNames) - 1; i >= 0; i-- {
		_, err := rdb.Exec("DROP TABLE IF EXISTS " + db.AllTableNames[i])
		if err != nil {
			return errors.Newf("drop %s: %v", db.AllTableNames[i], err)
		}
	}
	return nil
}

func TestNewRoach(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	tt := []struct {
		name   string
		opts   []db.Option
		expErr bool
	}{
		{
			name: "valid",
			opts: []db.Option{
				db.WithDBName(conf.DBName),
				db.WithDSN(conf.FormatDSN()),
			},
			expErr: false,
		},
		{
			name:   "valid (no options)",
			opts:   nil,
			expErr: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := db.NewRoach(tc.opts...)
			if r == nil {
				t.Fatalf("Got nil roach")
			}
		})
	}
}

func TestRoach_InitDBIfNot(t *testing.T) {

	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	rdb := getDB(t, conf)
	defer rdb.Close()
	if err := r.InitDBIfNot(); err != nil {
		t.Fatalf("Initial init call failed: %v", err)
	}

	tt := []struct {
		name       string
		hasVersion bool
		version    []byte
		expErr     bool
	}{
		{
			name:       "first use",
			hasVersion: false,
			expErr:     false,
		},
		{
			name:       "versions equal",
			hasVersion: true,
			version:    []byte(strconv.Itoa(db.Version)),
			expErr:     false,
		},
		{
			name:       "db version smaller",
			hasVersion: true,
			version:    []byte(strconv.Itoa(db.Version - 1)),
			expErr:     true,
		},
		{
			name:       "db version bigger",
			hasVersion: true,
			version:    []byte(strconv.Itoa(db.Version + 1)),
			expErr:     true,
		},
	}

	cols := db.ColDesc(db.ColKey, db.ColValue, db.ColUpdateDate)
	updCols := db.ColDesc(db.ColValue, db.ColUpdateDate)
	upsertQ := `
		INSERT INTO ` + db.TblConfigurations + ` (` + cols + `)
			VALUES ('db.version', $1, CURRENT_TIMESTAMP)
			ON CONFLICT (` + db.ColKey + `)
			DO UPDATE SET (` + updCols + `) = ($1, CURRENT_TIMESTAMP)`
	delQ := `DELETE FROM ` + db.TblConfigurations + ` WHERE ` + db.ColKey + `='db.version'`

	for _, tc := range tt {
		if _, err := rdb.Exec(delQ); err != nil {
			t.Fatalf("Error setting up: clear previous config: %v", err)
		}
		if tc.hasVersion {
			if _, err := rdb.Exec(upsertQ, tc.version); err != nil {
				t.Fatalf("Error setting up: insert test config: %v", err)
			}
		}
		t.Run(tc.name, func(t *testing.T) {
			r = newRoach(t, conf)
			err := r.InitDBIfNot()
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error, got nil")
				}
				// set db to have correct version (init error should be cached not queried)
				if _, err := rdb.Exec(upsertQ, []byte(strconv.Itoa(db.Version))); err != nil {
					t.Fatalf("Error setting up: insert test config: %v", err)
				}
				if err := r.InitDBIfNot(); err == nil {
					t.Fatalf("Subsequent init db not returning error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got an error: %v", err)
			}
			// set db to have incorrect version (isInit flag should be cached, not queried)
			if _, err := rdb.Exec(upsertQ, []byte(strconv.Itoa(db.Version+10))); err != nil {
				t.Fatalf("Error setting up: insert test config: %v", err)
			}
			if err = r.InitDBIfNot(); err != nil {
				t.Fatalf("Subsequent init not working")
			}
		})
	}
}

func newRoach(t *testing.T, conf crdb.Config) *db.Roach {
	r := db.NewRoach(
		db.WithDBName(conf.DBName),
		db.WithDSN(conf.FormatDSN()),
	)
	if r == nil {
		t.Fatalf("Got nil roach")
	}
	return r
}

func getDB(t *testing.T, conf crdb.Config) *sql.DB {
	DB, err := sql.Open("postgres", conf.FormatDSN())
	if err != nil {
		t.Fatalf("new db instance: %s", err)
	}
	return DB
}
