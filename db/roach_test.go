package db_test

import (
	"testing"

	"database/sql"

	"strconv"

	"github.com/tomogoma/authms/db"
	testingH "github.com/tomogoma/authms/testing"
	"github.com/tomogoma/go-commons/database/cockroach"
)

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
				db.WithDBName(conf.DBName()),
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
	upsertQ := `UPSERT INTO ` + db.TblConfigurations + ` (` + cols + `)
					VALUES ('db.version', $1, CURRENT_TIMESTAMP)`
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

func newRoach(t *testing.T, conf cockroach.DSN) *db.Roach {
	r := db.NewRoach(
		db.WithDBName(conf.DBName()),
		db.WithDSN(conf.FormatDSN()),
	)
	if r == nil {
		t.Fatalf("Got nil roach")
	}
	return r
}

func getDB(t *testing.T, conf cockroach.DSN) *sql.DB {
	db, err := cockroach.DBConn(conf)
	if err != nil {
		t.Fatalf("unable to tear down: cockroach.DBConn(): %s", err)
	}
	return db
}

func setup(t *testing.T) cockroach.DSN {
	conf := testingH.ReadConfig(t)
	conf.Database.DB = conf.Database.DB + "_test"
	return conf.Database
}

func tearDown(t *testing.T, conf cockroach.DSN) {
	rdb := getDB(t, conf)
	_, err := rdb.Exec("DROP DATABASE IF EXISTS " + conf.DBName())
	if err != nil {
		t.Fatalf("unable to tear down db: %s", err)
	}
}
