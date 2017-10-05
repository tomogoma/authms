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
