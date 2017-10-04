package db_test

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/model"
	testingH "github.com/tomogoma/authms/testing"
	"github.com/tomogoma/go-commons/database/cockroach"
)

func setup(t *testing.T) cockroach.DSN {
	conf := testingH.ReadConfig(t)
	conf.Database.DB = conf.Database.DB + "_test"
	return conf.Database
}

func tearDown(t *testing.T, conf cockroach.DSN) {
	rdb := getDB(t, conf)
	q := "SHOW TABLES FROM " + conf.DBName()
	rows , err := rdb.Query(q)
	if err != nil {
		t.Logf("Error tearing down: show tables: %v", err)
	}
	defer rows.Close()
	var tblNm string
	for rows.Next() {
		if err = rows.Scan(&tblNm); err != nil {
			t.Errorf("Error tearing down: show tables: scan row: %v", err)
		}
		_, err := rdb.Exec("DELETE FROM " + tblNm)
		if err != nil {
			t.Fatalf("Error tearing down: delete table: %v", err)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Error tearing down: show tables: iterating rows: %v", err)
	}
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

func TestRoach_InsertGroup(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	tt := []struct {
		testName string
		grpName  string
		acl      int
		expErr   bool
	}{
		{testName: "valid", grpName: "firstGroupName", acl: 5, expErr: false},
		{testName: "valid min acl", grpName: "new group 1", acl: 0, expErr: false},
		{testName: "valid max acl", grpName: "new group 2", acl: 10, expErr: false},
		{testName: "exists", grpName: "firstGroupName", acl: 7, expErr: true},
		{testName: "empty name", grpName: "", acl: 2, expErr: true},
		{testName: "acl too big", grpName: "new group 3", acl: 11, expErr: true},
		{testName: "acl too small", grpName: "new group 4", acl: -1, expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			grp, err := r.InsertGroup(tc.grpName, tc.acl)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if grp == nil {
				t.Fatalf("Got nil group")
			}
			if grp.ID == "" {
				t.Errorf("ID was not assigned")
			}
			if grp.UpdateDate.Before(time.Now().Add(-1 * time.Minute)) {
				t.Errorf("UpdateDate was not assigned")
			}
			if grp.CreateDate.Before(time.Now().Add(-1 * time.Minute)) {
				t.Errorf("CreateDate was not assigned")
			}
			if grp.AccessLevel != tc.acl {
				t.Errorf("AccessLevel mismatch, expect %d, got %d",
					tc.acl, grp.AccessLevel)
			}
			if grp.Name != tc.grpName {
				t.Errorf("Name mismatch, expect %d, got %d",
					tc.grpName, grp.Name)
			}
		})
	}
}

func TestRoach_InsertUserType(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	tt := []struct {
		testName string
		utName   string
		expErr   bool
	}{
		{testName: "valid", utName: "firstName", expErr: false},
		{testName: "empty name", utName: "", expErr: true},
		{testName: "repeated name", utName: "firstName", expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			ret, err := r.InsertUserType(tc.utName)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if ret == nil {
				t.Fatalf("Got nil group")
			}
			if ret.ID == "" {
				t.Errorf("ID was not assigned")
			}
			if ret.UpdateDate.Before(time.Now().Add(-1 * time.Minute)) {
				t.Errorf("UpdateDate was not assigned")
			}
			if ret.CreateDate.Before(time.Now().Add(-1 * time.Minute)) {
				t.Errorf("CreateDate was not assigned")
			}
			if ret.Name != tc.utName {
				t.Errorf("Name mismatch, expect %s, got %s",
					tc.utName, ret.Name)
			}
		})
	}
}

func TestRoach_InsertUserAtomic(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	ut, err := r.InsertUserType("test")
	if err != nil {
		t.Fatalf("Error setting up: insert user type: %v", err)
	}
	tt := []struct {
		testName string
		ut       model.UserType
		password []byte
		expErr   bool
	}{
		{testName: "valid", ut: *ut, password: []byte("12345678"), expErr: false},
		{testName: "bad typeID", ut: model.UserType{ID: "invalid"}, password: []byte("12345678"), expErr: true},
		{testName: "short password", ut: *ut, password: []byte("1234567"), expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			r.ExecuteTx(func(tx *sql.Tx) error {
				ret, err := r.InsertUserAtomic(tx, tc.ut, tc.password)
				if tc.expErr {
					if err == nil {
						t.Fatalf("Expected an error, got nil")
					}
					return nil
				}
				if err != nil {
					t.Fatalf("Got error: %v", err)
				}
				if ret == nil {
					t.Fatalf("Got nil group")
				}
				if ret.ID == "" {
					t.Errorf("ID was not assigned")
				}
				if ret.UpdateDate.Before(time.Now().Add(-1 * time.Minute)) {
					t.Errorf("UpdateDate was not assigned")
				}
				if ret.CreateDate.Before(time.Now().Add(-1 * time.Minute)) {
					t.Errorf("CreateDate was not assigned")
				}
				if ret.Type != tc.ut {
					t.Errorf("User type mismatch, expect %+v, got %+v",
						tc.ut, ret.Type)
				}
				return nil
			})
		})
	}
}

func TestRoach_InsertUserDeviceAtomic(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	tt := []struct {
		testName string
		devID    string
		usrID    string
		expErr   bool
	}{
		{testName: "valid", devID: "a-dev-id", usrID: usr.ID, expErr: false},
		{testName: "bad user ID", devID: "a-dev-id", usrID: "bad id", expErr: true},
		{testName: "empty dev ID", devID: "", usrID: "bad id", expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			r.ExecuteTx(func(tx *sql.Tx) error {
				ret, err := r.InsertUserDeviceAtomic(tx, tc.usrID, tc.devID)
				if tc.expErr {
					if err == nil {
						t.Fatalf("Expected an error, got nil")
					}
					return nil
				}
				if err != nil {
					t.Fatalf("Got error: %v", err)
				}
				if ret == nil {
					t.Fatalf("Got nil group")
				}
				if ret.ID == "" {
					t.Errorf("ID was not assigned")
				}
				if ret.UpdateDate.Before(time.Now().Add(-1 * time.Minute)) {
					t.Errorf("UpdateDate was not assigned")
				}
				if ret.CreateDate.Before(time.Now().Add(-1 * time.Minute)) {
					t.Errorf("CreateDate was not assigned")
				}
				if ret.UserID != tc.usrID {
					t.Errorf("User ID mismatch, expect %s, got %s",
						tc.usrID, ret.UserID)
				}
				if ret.DeviceID != tc.devID {
					t.Errorf("Device ID mismatch, expect %s, got %s",
						tc.devID, ret.DeviceID)
				}
				return nil
			})
		})
	}
}

// TestRoach_InsertUserNameAtomic shares test cases with TestRoach_InsertUserName
// because they use the same underlying implementation.
func TestRoach_InsertUserNameAtomic(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	r.ExecuteTx(func(tx *sql.Tx) error {
		ret, err := r.InsertUserNameAtomic(tx, usr.ID, "a-username")
		if err != nil {
			t.Fatalf("Got error: %v", err)
		}
		if ret == nil {
			t.Fatalf("Got nil group")
		}
		if ret.ID == "" {
			t.Errorf("ID was not assigned")
		}
		if ret.UpdateDate.Before(time.Now().Add(-1 * time.Minute)) {
			t.Errorf("UpdateDate was not assigned")
		}
		if ret.CreateDate.Before(time.Now().Add(-1 * time.Minute)) {
			t.Errorf("CreateDate was not assigned")
		}
		if ret.UserID != usr.ID {
			t.Errorf("User ID mismatch, expect %s, got %s",
				usr.ID, ret.UserID)
		}
		if ret.Value != "a-username" {
			t.Errorf("Username mismatch, expect a-username, got %s", ret.Value)
		}
		return nil
	})
}

func TestRoach_InsertUserName(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	tt := []struct {
		testName string
		username string
		usrID    string
		expErr   bool
	}{
		{testName: "valid", username: "a-username", usrID: usr.ID, expErr: false},
		{testName: "bad user ID", username: "a-dev-id", usrID: "bad id", expErr: true},
		{testName: "empty username", username: "", usrID: "bad id", expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			r.ExecuteTx(func(tx *sql.Tx) error {
				ret, err := r.InsertUserNameAtomic(tx, tc.usrID, tc.username)
				if tc.expErr {
					if err == nil {
						t.Fatalf("Expected an error, got nil")
					}
					return nil
				}
				if err != nil {
					t.Fatalf("Got error: %v", err)
				}
				if ret == nil {
					t.Fatalf("Got nil group")
				}
				if ret.ID == "" {
					t.Errorf("ID was not assigned")
				}
				if ret.UpdateDate.Before(time.Now().Add(-1 * time.Minute)) {
					t.Errorf("UpdateDate was not assigned")
				}
				if ret.CreateDate.Before(time.Now().Add(-1 * time.Minute)) {
					t.Errorf("CreateDate was not assigned")
				}
				if ret.UserID != tc.usrID {
					t.Errorf("User ID mismatch, expect %s, got %s",
						tc.usrID, ret.UserID)
				}
				if ret.Value != tc.username {
					t.Errorf("Username mismatch, expect %s, got %s",
						tc.username, ret.Value)
				}
				return nil
			})
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

func insertUser(t *testing.T, r *db.Roach) *model.User {
	ut, err := r.InsertUserType("test")
	if err != nil {
		t.Fatalf("Error setting up: insert user type: %v", err)
	}
	var usr *model.User
	err = r.ExecuteTx(func(tx *sql.Tx) error {
		usr, err = r.InsertUserAtomic(tx, *ut, []byte("CsH359UP"))
		return err
	})
	if err != nil {
		t.Fatalf("Error setting up: insert user: %v", err)
	}
	return usr
}
