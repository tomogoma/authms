package db_test

import (
	"testing"
	"time"

	"reflect"

	"github.com/pborman/uuid"
	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/model"
)

var currGrpACL = float32(0.0)

func TestRoach_InsertGroup(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	tt := []struct {
		testName string
		grpName  string
		acl      float32
		expErr   bool
	}{
		{testName: "valid", grpName: "firstGroupName", acl: 5.565, expErr: false},
		{testName: "valid min acl", grpName: "new group 1", acl: 0, expErr: false},
		{testName: "valid max acl", grpName: "new group 2", acl: 10, expErr: false},
		{testName: "exists", grpName: "firstGroupName", acl: 7, expErr: true},
		{testName: "empty name", grpName: "", acl: 2, expErr: true},
		{testName: "acl too big", grpName: "new group 3", acl: 11, expErr: true},
		{testName: "acl too small", grpName: "new group 4", acl: -1, expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			grp, err := r.InsertGroup(tc.grpName, tc.acl)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if grp == nil {
				t.Fatalf("Got nil group")
			}
			if grp.ID == "" {
				t.Errorf("ID was not assigned")
			}
			if grp.UpdateDate.Before(time.Now().Add(-1 * time.Minute)) {
				t.Errorf("UpdateDate was not assigned")
			}
			if grp.CreateDate.Before(time.Now().Add(-1 * time.Minute)) {
				t.Errorf("CreateDate was not assigned")
			}
			if grp.AccessLevel != tc.acl {
				t.Errorf("AccessLevel mismatch, expect %d, got %d",
					tc.acl, grp.AccessLevel)
			}
			if grp.Name != tc.grpName {
				t.Errorf("Name mismatch, expect %d, got %d",
					tc.grpName, grp.Name)
			}
		})
	}
}

func TestRoach_Group(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expGrp := insertGroup(t, r)
	tt := []struct {
		name        string
		grpID       string
		expNotFound bool
	}{
		{name: "found", grpID: expGrp.ID, expNotFound: false},
		{name: "not found", grpID: "123", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actGrp, err := r.Group(tc.grpID)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expGrp, actGrp) {
				t.Fatalf("Group mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expGrp, actGrp)
			}
		})
	}
}

func TestRoach_GroupByName(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expGrp := insertGroup(t, r)
	tt := []struct {
		name        string
		grpName     string
		expNotFound bool
	}{
		{name: "found", grpName: expGrp.Name, expNotFound: false},
		{name: "not found", grpName: "none-exist-name", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actGrp, err := r.GroupByName(tc.grpName)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expGrp, actGrp) {
				t.Fatalf("Group mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expGrp, actGrp)
			}
		})
	}
}

func TestRoach_GroupsByUserID(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	grp1 := insertGroup(t, r)
	grp2 := insertGroup(t, r)
	usr := insertUser(t, r)
	expGrps := []model.Group{*grp1, *grp2}
	addUserToGroups(t, r, usr.ID, grp1.ID, grp2.ID)
	tt := []struct {
		name        string
		usrID       string
		expNotFound bool
	}{
		{name: "found", usrID: usr.ID, expNotFound: false},
		{name: "not found", usrID: "123", expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actGrps, err := r.GroupsByUserID(tc.usrID)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expGrps, actGrps) {
				t.Fatalf("Group mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expGrps, actGrps)
			}
		})
	}
}

func insertGroup(t *testing.T, r *db.Roach) *model.Group {
	grp, err := r.InsertGroup(uuid.New(), currGrpACL)
	if err != nil {
		t.Fatalf("Error setting up: insert group: %v", err)
	}
	currGrpACL += 0.1
	return grp
}
