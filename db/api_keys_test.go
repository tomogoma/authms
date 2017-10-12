package db_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/api"
)

func TestRoach_InsertAPIKey(t *testing.T) {
	setupTime := time.Now()
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	validKey := strings.Repeat("axui", 14)
	tt := []struct {
		testName string
		key      string
		usrID    string
		expErr   bool
	}{
		{testName: "valid", key: validKey, usrID: usr.ID, expErr: false},
		{testName: "bad user ID", key: validKey, usrID: "bad id", expErr: true},
		{testName: "empty dev ID", key: "", usrID: usr.ID, expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			ret, err := r.InsertAPIKey(tc.usrID, tc.key)
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
			if ret.UpdateDate.Before(setupTime) {
				t.Errorf("UpdateDate was not assigned")
			}
			if ret.CreateDate.Before(setupTime) {
				t.Errorf("CreateDate was not assigned")
			}
			if ret.UserID != tc.usrID {
				t.Errorf("User ID mismatch, expect %s, got %s",
					tc.usrID, ret.UserID)
			}
			if ret.APIKey != tc.key {
				t.Errorf("API key mismatch, expect %s, got %s",
					tc.key, ret.APIKey)
			}
			return
		})
	}
}

func TestRoach_APIKeysByUserID(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	usrNoKeys := insertUser(t, r)
	key1 := insertAPIKey(t, r, usr.ID)
	key2 := insertAPIKey(t, r, usr.ID)
	expKeys := []api.Key{*key2, *key1}
	tt := []struct {
		name        string
		userID      string
		expNotFound bool
	}{
		{name: "found", userID: usr.ID, expNotFound: false},
		{name: "not found", userID: usrNoKeys.ID, expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actKeys, err := r.APIKeysByUserID(tc.userID, 0, 2)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expKeys, actKeys) {
				t.Errorf("API Keys mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expKeys, actKeys)
			}
		})
	}
}

func insertAPIKey(t *testing.T, r *db.Roach, usrID string) *api.Key {
	k, err := r.InsertAPIKey(usrID, strings.Repeat("x", 56))
	if err != nil {
		t.Fatalf("Error setting up: insert API key: %v", err)
	}
	return k
}
