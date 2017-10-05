package db_test

import (
	"time"
	"testing"
)

func TestRoach_InsertGroup(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	tt := []struct {
		testName string
		grpName  string
		acl      int
		expErr   bool
	}{
		{testName: "valid", grpName: "firstGroupName", acl: 5, expErr: false},
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