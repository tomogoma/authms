package dbhelper_test

import (
	"testing"
	"github.com/tomogoma/authms/proto/authms"
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
				AccessType: "LOGIN",
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
				AccessType: "LOGIN",
				SuccessStatus: true,
				DevID: "test-app-id",
			},
		},
		{
			Desc: "Empty access type",
			ExpErr: true,
			Hist: &authms.History{
				UserID: usr.ID,
				IpAddress: "127.0.0.1",
				AccessType: "",
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
	type TestCase struct {
		Desc string
		Hist *authms.History
	}
	tcs := []TestCase{
		{Desc: "All values provided", Hist: completeHistory(1)},
		{
			Desc: "Missing devID, IP Addr",
			Hist: &authms.History{
				AccessType: "LOGIN",
				SuccessStatus: true,
			},
		},
	}
	for _, tc := range tcs {
		func() {
			setUp(t)
			defer tearDown(t)
			m := newModel(t)
			usr := completeUser()
			insertUser(usr, t)
			q := `INSERT INTO history (userID, accessMethod,
			 		successful, devID, ipAddress, date)
				VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP())`
			db := getDB(t)
			tc.Hist.UserID = usr.ID
			_, err := db.Exec(q, tc.Hist.UserID, tc.Hist.AccessType,
				tc.Hist.SuccessStatus, tc.Hist.DevID, tc.Hist.IpAddress)
			if err != nil {
				t.Errorf("%s - Error setting up" +
					" (inserting history): %s", tc.Desc, err)
				return
			}
			offset := 0
			count := 1
			hists, err := m.GetHistory(tc.Hist.UserID, offset,
				count, tc.Hist.AccessType)
			if err != nil {
				t.Errorf("%s - model.GetHistory(): %s", tc.Desc, err)
				return
			}
			if len(hists) != 1 {
				t.Errorf("%s - Expected 1 history entry but got %d",
					tc.Desc, len(hists))
				return
			}
			if !compareHistory(tc.Hist, hists[0]) {
				t.Errorf("%s\nExpected:\t%+v\nGot:\t\t%+v",
					tc.Desc, tc.Hist, hists[0])
			}
		}()
	}
}

func completeHistory(userID int64) *authms.History {
	return &authms.History{
		UserID: userID,
		IpAddress: "127.0.0.1",
		AccessType: "LOGIN",
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