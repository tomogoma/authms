package db_test

import (
	"database/sql"
	"testing"
	"time"
)

func TestRoach_InsertUserDeviceAtomic_nilTx(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	_, err := r.InsertUserDeviceAtomic(nil, usr.ID, "a-dev-id-0")
	if err == nil {
		t.Errorf("(nil tx) - expected an error, got nil")
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
