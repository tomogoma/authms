package db_test

import (
	"database/sql"
	"testing"
	"time"
	"reflect"
	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/model"
)

func TestRoach_InsertUserNameAtomic_nilTx(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	_, err := r.InsertUserNameAtomic(nil, usr.ID, "a-username")
	if err == nil {
		t.Errorf("(nil tx) - expected an error, got nil")
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

func TestRoach_UpdateUsername(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	expUsrnm := insertUsername(t, r, usr.ID)
	expUsrnm.Value = "new-username"
	tt := []struct {
		name        string
		userID      string
		newUsrName  string
		expNotFound bool
	}{
		{name: "valid", userID: usr.ID, newUsrName: expUsrnm.Value, expNotFound: false},
		{name: "not found", userID: "123", newUsrName: expUsrnm.Value, expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			nun, err := r.UpdateUsername(tc.userID, tc.newUsrName)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Errorf("Expected IsNotFound, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if nun.UpdateDate.Equal(expUsrnm.CreateDate) || nun.UpdateDate.Before(expUsrnm.CreateDate) {
				t.Errorf("Update date not set correctly before/equal to create date")
			}
			expUsrnm.UpdateDate = nun.UpdateDate
			if !reflect.DeepEqual(expUsrnm, nun) {
				t.Errorf("Username mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expUsrnm, nun)
			}
		})
	}
}

func insertUsername(t *testing.T, r *db.Roach, usrID string) *model.Username {
	un, err := r.InsertUserName(usrID, uuid.New())
	if err != nil {
		t.Fatalf("Error setting up: insert username: %v", err)
	}
	return un
}
