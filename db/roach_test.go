package db_test

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	"bytes"

	"strings"

	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/model"
	testingH "github.com/tomogoma/authms/testing"
	"github.com/tomogoma/go-commons/database/cockroach"
)

var isInit bool

func setup(t *testing.T) cockroach.DSN {
	conf := testingH.ReadConfig(t)
	conf.Database.DB = conf.Database.DB + "_test"
	if !isInit {
		rdb := getDB(t, conf.Database)
		_, err := rdb.Exec("DROP DATABASE IF EXISTS " + conf.Database.DB)
		if err != nil {
			t.Fatalf("Error setting up: deleting db: %v", err)
		}
		isInit = true
	}
	return conf.Database
}

func tearDown(t *testing.T, conf cockroach.DSN) {
	rdb := getDB(t, conf)
	if _, err := rdb.Exec("SET DATABASE=" + conf.DB); err != nil {
		return
	}
	for i := len(db.AllTableNames) - 1; i >= 0; i-- {
		_, err := rdb.Exec("DELETE FROM " + db.AllTableNames[i])
		if err != nil {
			t.Fatalf("Error tearing down: delete %s: %v",
				db.AllTableNames[i], err)
		}
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
		{testName: "empty dev ID", devID: "", usrID: usr.ID, expErr: true},
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
		{testName: "empty username", username: "", usrID: usr.ID, expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			ret, err := r.InsertUserName(tc.usrID, tc.username)
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
			if ret.UserID != tc.usrID {
				t.Errorf("User ID mismatch, expect %s, got %s",
					tc.usrID, ret.UserID)
			}
			if ret.Value != tc.username {
				t.Errorf("Username mismatch, expect %s, got %s",
					tc.username, ret.Value)
			}
			return
		})
	}
}

// TestRoach_InsertUserPhoneAtomic shares test cases with TestRoach_InsertUserPhone
// because they use the same underlying implementation.
func TestRoach_InsertUserPhoneAtomic(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	r.ExecuteTx(func(tx *sql.Tx) error {
		ret, err := r.InsertUserPhoneAtomic(tx, usr.ID, "+254712345678", false)
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
		if ret.Address != "+254712345678" {
			t.Errorf("Phone number mismatch, expect +254712345678, got %s",
				ret.Address)
		}
		if ret.Verified != false {
			t.Errorf("verified mismatch, expect false, got %t", ret.Verified)
		}
		return nil
	})
}

func TestRoach_InsertUserPhone(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	tt := []struct {
		testName string
		usrID    string
		addr     string
		verified bool
		expErr   bool
	}{
		{testName: "valid", usrID: usr.ID, addr: "+254712345678", verified: true, expErr: false},
		{testName: "bad user ID", usrID: "bad id", addr: "+254712345678", verified: true, expErr: true},
		{testName: "empty phone", usrID: usr.ID, addr: "", expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			ret, err := r.InsertUserPhone(tc.usrID, tc.addr, tc.verified)
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
			if ret.UserID != tc.usrID {
				t.Errorf("User ID mismatch, expect %s, got %s",
					tc.usrID, ret.UserID)
			}
			if ret.Address != tc.addr {
				t.Errorf("address mismatch, expect %s, got %s",
					tc.addr, ret.Address)
			}
			if ret.Verified != tc.verified {
				t.Errorf("verified mismatch, expect %t, got %t",
					tc.verified, ret.Verified)
			}
			return
		})
	}
}

// TestRoach_InsertUserPhoneAtomic shares test cases with TestRoach_InsertUserPhone
// because they use the same underlying implementation.
func TestRoach_InsertPhoneTokenAtomic(t *testing.T) {
	setupTime := time.Now()
	dbt := []byte(strings.Repeat("x", 57))
	isUsed := false
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	phn := insertPhone(t, r, usr.ID)
	r.ExecuteTx(func(tx *sql.Tx) error {
		ret, err := r.InsertPhoneTokenAtomic(tx, usr.ID, phn.Address, dbt, isUsed, setupTime)
		if err != nil {
			t.Fatalf("Got error: %v", err)
		}
		if ret == nil {
			t.Fatalf("Got nil group")
		}
		if ret.ID == "" {
			t.Errorf("ID was not assigned")
		}
		if ret.IssueDate.Before(setupTime) {
			t.Errorf("Issue date was not assigned")
		}
		if ret.UserID != usr.ID {
			t.Errorf("User ID mismatch, expect %s, got %s",
				usr.ID, ret.UserID)
		}
		if ret.Address != phn.Address {
			t.Errorf("Invalid phone: expect %s, got %s", phn.Address, ret.Address)
		}
		if !bytes.Equal(ret.Token, dbt) {
			t.Errorf("Invalid db token: expect %s, got %s", dbt, ret.Token)
		}
		if ret.IsUsed != isUsed {
			t.Errorf("Invalid used val: expect %t, got %t", isUsed, ret.IsUsed)
		}
		if ret.ExpiryDate != setupTime {
			t.Errorf("Invalid expiry: expect %v, got %v", setupTime, ret.ExpiryDate)
		}
		return nil
	})
}

func TestRoach_InsertPhoneToken(t *testing.T) {
	setUpTime := time.Now()
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	phn := insertPhone(t, r, usr.ID)
	validDBT := []byte(strings.Repeat("x", 57))
	tt := []struct {
		testName string
		usrID    string
		addr     string
		dbt      []byte
		isUsed   bool
		expiry   time.Time
		expErr   bool
	}{
		{
			testName: "valid",
			usrID:    usr.ID,
			addr:     phn.Address,
			dbt:      validDBT,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   false,
		},
		{
			testName: "bad user id",
			usrID:    "bad id",
			addr:     phn.Address,
			dbt:      validDBT,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
		{
			testName: "empty phone",
			usrID:    usr.ID,
			addr:     "",
			dbt:      validDBT,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
		{
			testName: "bad phone",
			usrID:    usr.ID,
			addr:     "bad phone",
			dbt:      validDBT,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
		{
			testName: "empty dbt",
			usrID:    usr.ID,
			addr:     phn.Address,
			dbt:      []byte{},
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
		{
			testName: "nil dbt",
			usrID:    usr.ID,
			addr:     phn.Address,
			dbt:      nil,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			ret, err := r.InsertPhoneToken(tc.usrID, tc.addr, tc.dbt, tc.isUsed, tc.expiry)
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
			if ret.IssueDate.Before(setUpTime) {
				t.Errorf("Issue date was not assigned")
			}
			if ret.UserID != tc.usrID {
				t.Errorf("User ID mismatch, expect %s, got %s",
					tc.usrID, ret.UserID)
			}
			if ret.Address != tc.addr {
				t.Errorf("Invalid phone: expect %s, got %s", tc.addr, ret.Address)
			}
			if !bytes.Equal(ret.Token, tc.dbt) {
				t.Errorf("Invalid db token: expect %s, got %s", tc.dbt, ret.Token)
			}
			if ret.IsUsed != tc.isUsed {
				t.Errorf("Invalid used val: expect %t, got %t", tc.isUsed, ret.IsUsed)
			}
			if ret.ExpiryDate != tc.expiry {
				t.Errorf("Invalid expiry: expect %v, got %v", tc.expiry, ret.ExpiryDate)
			}
			return
		})
	}
}

// TestRoach_InsertUserEmailAtomic shares test cases with TestRoach_InsertUserEmail
// because they use the same underlying implementation.
func TestRoach_InsertUserEmailAtomic(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	r.ExecuteTx(func(tx *sql.Tx) error {
		ret, err := r.InsertUserEmailAtomic(tx, usr.ID, "test@mailinator.com", false)
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
		if ret.Address != "test@mailinator.com" {
			t.Errorf("Address mismatch, expect test@mailinator.com, got %s",
				ret.Address)
		}
		if ret.Verified != false {
			t.Errorf("verified mismatch, expect false, got %t", ret.Verified)
		}
		return nil
	})
}

func TestRoach_InsertUserEmail(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	tt := []struct {
		testName string
		usrID    string
		addr     string
		verified bool
		expErr   bool
	}{
		{testName: "valid", usrID: usr.ID, addr: "test@mailinator.com", verified: true, expErr: false},
		{testName: "bad user ID", usrID: "bad id", addr: "test@mailinator.com", verified: true, expErr: true},
		{testName: "empty phone", usrID: usr.ID, addr: "", expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			ret, err := r.InsertUserEmail(tc.usrID, tc.addr, tc.verified)
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
			if ret.UserID != tc.usrID {
				t.Errorf("User ID mismatch, expect %s, got %s",
					tc.usrID, ret.UserID)
			}
			if ret.Address != tc.addr {
				t.Errorf("address mismatch, expect %s, got %s",
					tc.addr, ret.Address)
			}
			if ret.Verified != tc.verified {
				t.Errorf("verified mismatch, expect %t, got %t",
					tc.verified, ret.Verified)
			}
			return
		})
	}
}

func TestRoach_InsertUserFbIDAtomic(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	usr2 := insertUser(t, r)
	tt := []struct {
		testName string
		usrID    string
		fbID     string
		verified bool
		expErr   bool
	}{
		{testName: "valid", usrID: usr.ID, fbID: "an-fb-id-1", verified: false, expErr: false},
		{testName: "valid verified", usrID: usr2.ID, fbID: "an-fb-id-2", verified: true, expErr: false},
		{testName: "bad userID", usrID: "bad userID", fbID: "an-fb-id", verified: false, expErr: true},
		{testName: "empty fbID", usrID: usr.ID, fbID: "", verified: false, expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			r.ExecuteTx(func(tx *sql.Tx) error {
				ret, err := r.InsertUserFbIDAtomic(tx, tc.usrID, tc.fbID, tc.verified)
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
				if ret.FacebookID != tc.fbID {
					t.Errorf("Facebook ID mismatch, expect %s, got %s",
						tc.fbID, ret.FacebookID)
				}
				if ret.Verified != tc.verified {
					t.Errorf("Verified mismatch, expect %t, got %t",
						tc.verified, ret.Verified)
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
	ut, err := r.InsertUserType(uuid.New())
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

func insertPhone(t *testing.T, r *db.Roach, usrID string) *model.VerifLogin {
	phn, err := r.InsertUserPhone(usrID, "+254712345678", false)
	if err != nil {
		t.Fatalf("Error setting up: insert phone: %v", err)
	}
	return phn
}
