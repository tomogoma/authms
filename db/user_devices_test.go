package db_test

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/model"
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

func TestRoach_UserDevicesByUserID(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	usrNoDevs := insertUser(t, r)
	dev1 := insertUserDevice(t, r, usr.ID)
	dev2 := insertUserDevice(t, r, usr.ID)
	expDevs := []model.Device{*dev1, *dev2}
	tt := []struct {
		name        string
		userID      string
		expNotFound bool
	}{
		{name: "found", userID: usr.ID, expNotFound: false},
		{name: "not found", userID: usrNoDevs.ID, expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			devs, err := r.UserDevicesByUserID(tc.userID)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expDevs, devs) {
				t.Errorf("Devices mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expDevs, devs)
			}
		})
	}
}

func insertUserDevice(t *testing.T, r *db.Roach, usrID string) *model.Device {
	var dev *model.Device
	var err error
	err = r.ExecuteTx(func(tx *sql.Tx) error {
		dev, err = r.InsertUserDeviceAtomic(tx, usrID, uuid.New())
		return err
	})
	if err != nil {
		t.Fatalf("Error setting up: insert device: %v", err)
	}
	return dev
}
