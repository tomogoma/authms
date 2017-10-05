package db_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/model"
)

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
		nilTx    bool
		expErr   bool
	}{
		{testName: "valid", ut: *ut, password: []byte("12345678"), expErr: false},
		{testName: "bad typeID", ut: model.UserType{ID: "invalid"}, password: []byte("12345678"), expErr: true},
		{testName: "short password", ut: *ut, password: []byte("1234567"), expErr: true},
	}
	_, err = r.InsertUserAtomic(nil, *ut, []byte("123456789"))
	if err == nil {
		t.Errorf("(nil tx) - expected an error, got nil")
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			if tc.nilTx {
			}
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

func TestRoach_UpdatePassword(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	validPass := []byte("A g00d P@$$wo%d") //chars >= 8
	tt := []struct {
		name     string
		userID   string
		password []byte
		expErr   bool
	}{
		{name: "valid", userID: usr.ID, password: validPass, expErr: false},
		{name: "non-exist userID", userID: "not-exist", password: validPass, expErr: true},
		{name: "short password", userID: usr.ID, password: []byte("7 chars"), expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := r.UpdatePassword(tc.userID, tc.password)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got an error: %v", err)
			}
		})
	}
}

func TestRoach_UpdatePasswordAtomic(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	validPass := []byte("A g00d P@$$wo%d") //chars >= 8
	r.ExecuteTx(func(tx *sql.Tx) error {
		err := r.UpdatePasswordAtomic(tx, usr.ID, validPass)
		if err != nil {
			t.Fatalf("Got an error: %v", err)
		}
		return nil
	})
	err := r.UpdatePasswordAtomic(nil, usr.ID, validPass)
	if err == nil {
		t.Fatalf("nil tx - expected an error, got nil")
	}
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
