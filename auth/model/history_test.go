package model_test

import (
	"testing"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/auth/model"
)

func TestModel_SaveHistory(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	m := newModel(t)
	usr := completeUser()
	insertUser(usr, t)
	type SaveHistoryTestCase struct {
		Desc   string
		Hist   *authms.History
		ExpErr bool
	}
	tcs := []SaveHistoryTestCase{
		{
			Desc: "Valid history",
			ExpErr: false,
			Hist: completeHistory(usr.ID),
		},
		{
			Desc: "Invalid user ID",
			ExpErr: true,
			Hist: &authms.History{
				UserID: -1,
				IpAddress: "127.0.0.1",
				AccessType: model.AccessLogin,
				SuccessStatus: true,
				DevID: "test-app-id",
			},
		},
		{
			Desc: "Non existent user ID",
			ExpErr: true,
			Hist: &authms.History{
				UserID: 1,
				IpAddress: "127.0.0.1",
				AccessType: model.AccessLogin,
				SuccessStatus: true,
				DevID: "test-app-id",
			},
		},
		{
			Desc: "Invalid access type",
			ExpErr: true,
			Hist: &authms.History{
				UserID: usr.ID,
				IpAddress: "127.0.0.1",
				AccessType: "invalid-access-type",
				SuccessStatus: true,
				DevID: "test-app-id",
			},
		},
	}
	for _, tc := range tcs {
		func() {
			err := m.SaveHistory(tc.Hist)
			if tc.ExpErr {
				if err == nil {
					t.Errorf("%s - expected an error but got none",
						tc.Desc)
				}
				return
			} else if err != nil {
				t.Errorf("%s - model.SaveHistory(): %v",
					tc.Desc, err)
				return
			}
			q := `SELECT id, accessMethod, successful, userID, date, devID, ipAddress
				FROM history WHERE id=$1`
			db := getDB(t)
			hist := new(authms.History)
			err = db.QueryRow(q, tc.Hist.ID).Scan(&hist.ID,
				&hist.AccessType, &hist.SuccessStatus,
				&hist.UserID, &hist.Date, &hist.DevID,
				&hist.IpAddress)
			if err != nil {
				t.Fatalf("%s - Error fetching history for" +
					" validation: %s", tc.Desc, err)
				return
			}
			if !compareHistory(tc.Hist, hist) {
				t.Errorf("%s - expected %+v but got %+v",
					tc.Desc, tc.Hist, hist)
			}
		}()
	}
}

func TestModel_GetHistory(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	m := newModel(t)
	usr := completeUser()
	insertUser(usr, t)
	q := `INSERT INTO history (userID, accessMethod, successful, devID, ipAddress, date)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP())`
	db := getDB(t)
	hist := completeHistory(usr.ID)
	_, err := db.Exec(q, hist.UserID, hist.AccessType, hist.SuccessStatus, hist.DevID,
		hist.IpAddress)
	if err != nil {
		t.Fatalf("Error setting up (inserting history): %s", err)
	}
	offset := 0
	count := 1
	hists, err := m.GetHistory(hist.UserID, offset, count, hist.AccessType)
	if err != nil {
		t.Fatalf("model.GetHistory(): %s", err)
	}
	if len(hists) != 1 {
		t.Fatalf("Expected 1 history entry but got %d", len(hists))
	}
	if !compareHistory(hist, hists[0]) {
		t.Errorf("Expected %+v but got %+v", hist, hists[0])
	}
}

func completeHistory(userID int64) *authms.History {
	return &authms.History{
		UserID: userID,
		IpAddress: "127.0.0.1",
		AccessType: model.AccessLogin,
		SuccessStatus: true,
		DevID: "test-app-id",
	}
}

func compareHistory(act, exp *authms.History) bool {
	if exp == nil {
		return act != nil
	} else if act == nil {
		return false
	}
	if act.UserID != exp.UserID {
		return false
	}
	if act.IpAddress != exp.IpAddress {
		return false
	}
	if act.AccessType != exp.AccessType {
		return false
	}
	if act.SuccessStatus != exp.SuccessStatus {
		return false
	}
	if act.DevID != exp.DevID {
		return false
	}
	return true
}