package db_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/model"
)

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

func TestRoach_UserTypeByName(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expUt := insertUserType(t, r)
	tt := []struct {
		name        string
		utName      string
		expNotFound bool
	}{
		{name: "found", utName: expUt.Name, expNotFound: false},
		{name: "not found", utName: "not-exist", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actUt, err := r.UserTypeByName(tc.utName)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expUt, actUt) {
				t.Fatalf("UserType mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expUt, actUt)
			}
		})
	}
}

func insertUserType(t *testing.T, r *db.Roach) *model.UserType {
	ut, err := r.InsertUserType(uuid.New())
	if err != nil {
		t.Fatalf("Error setting up: insert usertype: %v", err)
	}
	return ut
}
