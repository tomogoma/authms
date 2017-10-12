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

var insUsrPass = []byte("CsH359UP")

func TestRoach_InsertUserAtomic_nilTx(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	ut, err := r.InsertUserType("test")
	if err != nil {
		t.Fatalf("Error setting up: insert user type: %v", err)
	}
	_, err = r.InsertUserAtomic(nil, *ut, []byte("123456789"))
	if err == nil {
		t.Errorf("(nil tx) - expected an error, got nil")
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

func TestRoach_UpdatePasswordAtomic_nilTx(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	validPass := []byte("A g00d P@$$wo%d") //chars >= 8
	err := r.UpdatePasswordAtomic(nil, usr.ID, validPass)
	if err == nil {
		t.Fatalf("nil tx - expected an error, got nil")
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
}

func TestRoach_AddUserToGroupAtomic_nilTx(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	grp := insertGroup(t, r)
	err := r.AddUserToGroupAtomic(nil, usr.ID, grp.ID)
	if err == nil {
		t.Errorf("(nil tx) - expected an error, got nil")
	}
}

func TestRoach_AddUserToGroupAtomic(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	grp := insertGroup(t, r)
	tt := []struct {
		testName string
		usrID    string
		grpID    string
		expErr   bool
	}{
		{testName: "valid", usrID: usr.ID, grpID: grp.ID, expErr: false},
		{testName: "none exist usrID", usrID: "123", grpID: grp.ID, expErr: true},
		{testName: "none exist grpID", usrID: usr.ID, grpID: "123", expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			r.ExecuteTx(func(tx *sql.Tx) error {
				err := r.AddUserToGroupAtomic(tx, tc.usrID, tc.grpID)
				if tc.expErr {
					if err == nil {
						t.Fatalf("Expected an error, got nil")
					}
					return nil
				}
				if err != nil {
					t.Fatalf("Got error: %v", err)
				}
				return nil
			})
		})
	}
}

func TestRoach_User(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expUsr := insertUser(t, r)
	grp1 := insertGroup(t, r)
	grp2 := insertGroup(t, r)
	addUserToGroups(t, r, expUsr.ID, grp1.ID, grp2.ID)
	expUsr.Groups = []model.Group{*grp1, *grp2}
	tt := []struct {
		name        string
		usrID       string
		expNotFound bool
	}{
		{name: "found", usrID: expUsr.ID, expNotFound: false},
		{name: "not found", usrID: "123", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actUsr, pass, err := r.User(tc.usrID)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expUsr, actUsr) {
				t.Fatalf("User mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expUsr, actUsr)
			}
			if !reflect.DeepEqual(insUsrPass, pass) {
				t.Fatalf("Password mismatch:\nExpect:\t%+v\nGot:\t%+v",
					insUsrPass, pass)
			}
		})
	}
}

func TestRoach_UserByDeviceID(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expUsr := insertUser(t, r)
	dev1 := insertUserDevice(t, r, expUsr.ID)
	dev2 := insertUserDevice(t, r, expUsr.ID)
	expUsr.Devices = []model.Device{*dev1, *dev2}
	tt := []struct {
		name        string
		devID       string
		expNotFound bool
	}{
		{name: "found", devID: dev1.DeviceID, expNotFound: false},
		{name: "not found", devID: "none-exist-dev-id", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actUsr, pass, err := r.UserByDeviceID(tc.devID)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expUsr, actUsr) {
				t.Fatalf("User mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expUsr, actUsr)
			}
			if !reflect.DeepEqual(insUsrPass, pass) {
				t.Fatalf("Password mismatch:\nExpect:\t%+v\nGot:\t%+v",
					insUsrPass, pass)
			}
		})
	}
}

func TestRoach_UserByEmail(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expUsr := insertUser(t, r)
	mail := insertEmail(t, r, expUsr.ID)
	expUsr.Email = *mail
	tt := []struct {
		name        string
		addr        string
		expNotFound bool
	}{
		{name: "found", addr: mail.Address, expNotFound: false},
		{name: "not found", addr: "bad@mail.", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actUsr, pass, err := r.UserByEmail(tc.addr)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expUsr, actUsr) {
				t.Fatalf("User mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expUsr, actUsr)
			}
			if !reflect.DeepEqual(insUsrPass, pass) {
				t.Fatalf("Password mismatch:\nExpect:\t%+v\nGot:\t%+v",
					insUsrPass, pass)
			}
		})
	}
}

func TestRoach_UserByPhone(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expUsr := insertUser(t, r)
	phone := insertPhone(t, r, expUsr.ID)
	expUsr.Phone = *phone
	tt := []struct {
		name        string
		num         string
		expNotFound bool
	}{
		{name: "found", num: phone.Address, expNotFound: false},
		{name: "not found", num: "+unknown-number.", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actUsr, pass, err := r.UserByPhone(tc.num)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expUsr, actUsr) {
				t.Fatalf("User mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expUsr, actUsr)
			}
			if !reflect.DeepEqual(insUsrPass, pass) {
				t.Fatalf("Password mismatch:\nExpect:\t%+v\nGot:\t%+v",
					insUsrPass, pass)
			}
		})
	}
}

func TestRoach_UserByUsername(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expUsr := insertUser(t, r)
	usrnm := insertUsername(t, r, expUsr.ID)
	expUsr.UserName = *usrnm
	tt := []struct {
		name        string
		usrName     string
		expNotFound bool
	}{
		{name: "found", usrName: usrnm.Value, expNotFound: false},
		{name: "not found", usrName: "unknown-usrname", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actUsr, pass, err := r.UserByUsername(tc.usrName)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expUsr, actUsr) {
				t.Fatalf("User mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expUsr, actUsr)
			}
			if !reflect.DeepEqual(insUsrPass, pass) {
				t.Fatalf("Password mismatch:\nExpect:\t%+v\nGot:\t%+v",
					insUsrPass, pass)
			}
		})
	}
}

func TestRoach_UserByFacebook(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expUsr := insertUser(t, r)
	fbID := insertFbID(t, r, expUsr.ID)
	expUsr.Facebook = *fbID
	tt := []struct {
		name        string
		fbID        string
		expNotFound bool
	}{
		{name: "found", fbID: fbID.FacebookID, expNotFound: false},
		{name: "not found", fbID: "unknown-fb-id", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actUsr, err := r.UserByFacebook(tc.fbID)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expUsr, actUsr) {
				t.Fatalf("User mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expUsr, actUsr)
			}
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
		usr, err = r.InsertUserAtomic(tx, *ut, insUsrPass)
		return err
	})
	if err != nil {
		t.Fatalf("Error setting up: insert user: %v", err)
	}
	return usr
}

func addUserToGroups(t *testing.T, r *db.Roach, usrID string, groupIDs ...string) {
	err := r.ExecuteTx(func(tx *sql.Tx) error {
		for _, grpID := range groupIDs {
			if err := r.AddUserToGroupAtomic(tx, usrID, grpID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Error setting up: add users to group: %v", err)
	}
}
