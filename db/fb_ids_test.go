package db_test

import (
	"database/sql"
	"testing"
	"time"
)

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
