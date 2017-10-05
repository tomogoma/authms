package db_test

import (
	"database/sql"
	"testing"
	"time"
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
