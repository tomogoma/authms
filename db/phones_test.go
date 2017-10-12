package db_test

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"
	"time"

	"reflect"

	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/model"
)

func TestRoach_InsertUserPhoneAtomic_nilTx(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	_, err := r.InsertUserPhoneAtomic(nil, usr.ID, "+254712345678", false)
	if err == nil {
		t.Errorf("(nil tx) - expected an error, got nil")
	}
}

// TestRoach_InsertUserPhoneAtomic shares test cases with TestRoach_InsertUserPhone
// because they use the same underlying implementation.
func TestRoach_InsertUserPhoneAtomic(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	r.ExecuteTx(func(tx *sql.Tx) error {
		ret, err := r.InsertUserPhoneAtomic(tx, usr.ID, "+254712345678", false)
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
		if ret.Address != "+254712345678" {
			t.Errorf("Phone number mismatch, expect +254712345678, got %s",
				ret.Address)
		}
		if ret.Verified != false {
			t.Errorf("verified mismatch, expect false, got %t", ret.Verified)
		}
		return nil
	})
}

func TestRoach_InsertUserPhone(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	tt := []struct {
		testName string
		usrID    string
		addr     string
		verified bool
		expErr   bool
	}{
		{testName: "valid", usrID: usr.ID, addr: "+254712345678", verified: true, expErr: false},
		{testName: "bad user ID", usrID: "bad id", addr: "+254712345678", verified: true, expErr: true},
		{testName: "empty phone", usrID: usr.ID, addr: "", expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			ret, err := r.InsertUserPhone(tc.usrID, tc.addr, tc.verified)
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
			if ret.Address != tc.addr {
				t.Errorf("address mismatch, expect %s, got %s",
					tc.addr, ret.Address)
			}
			if ret.Verified != tc.verified {
				t.Errorf("verified mismatch, expect %t, got %t",
					tc.verified, ret.Verified)
			}
			return
		})
	}
}

func TestRoach_UpdateUserPhone(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	expPhn := insertPhone(t, r, usr.ID)
	expPhn.Address = "new-phone-no"
	expPhn.Verified = true
	tt := []struct {
		name         string
		userID       string
		newAddr      string
		newVerStatus bool
		expNotFound  bool
	}{
		{name: "valid", userID: usr.ID, newAddr: expPhn.Address, newVerStatus: expPhn.Verified, expNotFound: false},
		{name: "not found", userID: "123", newAddr: expPhn.Address, newVerStatus: expPhn.Verified, expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			updPhn, err := r.UpdateUserPhone(tc.userID, tc.newAddr, tc.newVerStatus)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Errorf("Expected IsNotFound, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if updPhn.UpdateDate.Equal(expPhn.CreateDate) || updPhn.UpdateDate.Before(expPhn.CreateDate) {
				t.Errorf("Update date not set correctly before/equal to create date")
			}
			expPhn.UpdateDate = updPhn.UpdateDate
			if !reflect.DeepEqual(expPhn, updPhn) {
				t.Errorf("Phone mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expPhn, updPhn)
			}
		})
	}
}

func TestRoach_UpdateUserPhoneAtomic_nilTx(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	phn := insertPhone(t, r, usr.ID)
	_, err := r.UpdateUserPhoneAtomic(nil, usr.ID, phn.ID, true)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

func TestRoach_InsertPhoneTokenAtomic_nilTx(t *testing.T) {
	setupTime := time.Now()
	dbt := []byte(strings.Repeat("x", 57))
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	phn := insertPhone(t, r, usr.ID)
	_, err := r.InsertPhoneTokenAtomic(nil, usr.ID, phn.Address, dbt, false, setupTime)
	if err == nil {
		t.Errorf("(nil tx) - expected an error, got nil")
	}
}

// TestRoach_InsertPhoneTokenAtomic shares test cases with TestRoach_InsertPhoneToken
// because they use the same underlying implementation.
func TestRoach_InsertPhoneTokenAtomic(t *testing.T) {
	setupTime := time.Now()
	dbt := []byte(strings.Repeat("x", 57))
	isUsed := false
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	phn := insertPhone(t, r, usr.ID)
	r.ExecuteTx(func(tx *sql.Tx) error {
		ret, err := r.InsertPhoneTokenAtomic(tx, usr.ID, phn.Address, dbt, isUsed, setupTime)
		if err != nil {
			t.Fatalf("Got error: %v", err)
		}
		if ret == nil {
			t.Fatalf("Got nil group")
		}
		if ret.ID == "" {
			t.Errorf("ID was not assigned")
		}
		if ret.IssueDate.Before(setupTime) {
			t.Errorf("Issue date was not assigned")
		}
		if ret.UserID != usr.ID {
			t.Errorf("User ID mismatch, expect %s, got %s",
				usr.ID, ret.UserID)
		}
		if ret.Address != phn.Address {
			t.Errorf("Invalid phone: expect %s, got %s", phn.Address, ret.Address)
		}
		if !bytes.Equal(ret.Token, dbt) {
			t.Errorf("Invalid db token: expect %s, got %s", dbt, ret.Token)
		}
		if ret.IsUsed != isUsed {
			t.Errorf("Invalid used val: expect %t, got %t", isUsed, ret.IsUsed)
		}
		// TODO truncate tc.expiry appropriately before testing
		//if ret.ExpiryDate != setupTime {
		//	t.Errorf("Invalid expiry: expect %v, got %v", setupTime, ret.ExpiryDate)
		//}
		return nil
	})
}

func TestRoach_InsertPhoneToken(t *testing.T) {
	setUpTime := time.Now()
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	phn := insertPhone(t, r, usr.ID)
	validDBT := []byte(strings.Repeat("x", 57))
	tt := []struct {
		testName string
		usrID    string
		addr     string
		dbt      []byte
		isUsed   bool
		expiry   time.Time
		expErr   bool
	}{
		{
			testName: "valid",
			usrID:    usr.ID,
			addr:     phn.Address,
			dbt:      validDBT,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   false,
		},
		{
			testName: "bad user id",
			usrID:    "bad id",
			addr:     phn.Address,
			dbt:      validDBT,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
		{
			testName: "empty phone",
			usrID:    usr.ID,
			addr:     "",
			dbt:      validDBT,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
		{
			testName: "bad phone",
			usrID:    usr.ID,
			addr:     "bad phone",
			dbt:      validDBT,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
		{
			testName: "empty dbt",
			usrID:    usr.ID,
			addr:     phn.Address,
			dbt:      []byte{},
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
		{
			testName: "nil dbt",
			usrID:    usr.ID,
			addr:     phn.Address,
			dbt:      nil,
			isUsed:   false,
			expiry:   setUpTime.Add(5 * time.Minute),
			expErr:   true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			ret, err := r.InsertPhoneToken(tc.usrID, tc.addr, tc.dbt, tc.isUsed, tc.expiry)
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
			if ret.IssueDate.Before(setUpTime) {
				t.Errorf("Issue date was not assigned")
			}
			if ret.UserID != tc.usrID {
				t.Errorf("User ID mismatch, expect %s, got %s",
					tc.usrID, ret.UserID)
			}
			if ret.Address != tc.addr {
				t.Errorf("Invalid phone: expect %s, got %s", tc.addr, ret.Address)
			}
			if !bytes.Equal(ret.Token, tc.dbt) {
				t.Errorf("Invalid db token: expect %s, got %s", tc.dbt, ret.Token)
			}
			if ret.IsUsed != tc.isUsed {
				t.Errorf("Invalid used val: expect %t, got %t", tc.isUsed, ret.IsUsed)
			}
			// TODO truncate tc.expiry appropriately before testing
			//if ret.ExpiryDate != tc.expiry {
			//	t.Errorf("Invalid expiry: expect %v, got %v", tc.expiry, ret.ExpiryDate)
			//}
			return
		})
	}
}

func TestRoach_PhoneTokens(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	usr := insertUser(t, r)
	usrNoTkns := insertUser(t, r)
	phn := insertPhone(t, r, usr.ID)
	dbt1 := insertPhoneToken(t, r, usr.ID, phn.Address)
	dbt2 := insertPhoneToken(t, r, usr.ID, phn.Address)
	expDBTs := []model.DBToken{*dbt2, *dbt1}
	tt := []struct {
		name        string
		userID      string
		expNotFound bool
	}{
		{name: "found", userID: usr.ID, expNotFound: false},
		{name: "not found", userID: usrNoTkns.ID, expNotFound: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actDBTs, err := r.PhoneTokens(tc.userID, 0, 2)
			if tc.expNotFound {
				if !r.IsNotFoundError(err) {
					t.Fatalf("Expected not found error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !reflect.DeepEqual(expDBTs, actDBTs) {
				t.Errorf("DBTokens mismatch:\nExpect:\t%+v\nGot:\t%+v",
					expDBTs, actDBTs)
			}
		})
	}
}

func insertPhone(t *testing.T, r *db.Roach, usrID string) *model.VerifLogin {
	phn, err := r.InsertUserPhone(usrID, "+254712345678", false)
	if err != nil {
		t.Fatalf("Error setting up: insert phone: %v", err)
	}
	return phn
}

func insertPhoneToken(t *testing.T, r *db.Roach, usrID, phone string) *model.DBToken {
	phnTkn, err := r.InsertPhoneToken(
		usrID,
		phone,
		[]byte(strings.Repeat("x", 56)),
		false,
		time.Now().Add(5*time.Minute),
	)
	if err != nil {
		t.Fatalf("Error setting up: insert phone token: %v", err)
	}
	return phnTkn
}
