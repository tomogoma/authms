package history_test

import (
	"testing"
	"time"

	"bitbucket.org/tomogoma/auth-ms/auth/model/history"
)

const (
	userID    = 56789034561382845
	ipAddress = "127.0.0.1"
	acm       = history.LoginAccess
	refSrv    = "SOME-SERV-ID"
	referall  = "SOME-OTHER-SERV-ID"
)

type hist struct {
	id         int
	acm        int
	userID     int
	ipAddress  string
	date       time.Time
	forService string
	referral   string
	successful bool
}

func (ld *hist) explodeParams() (int, int, bool, time.Time, string, string, string) {
	return ld.userID, ld.acm, ld.successful, ld.date, ld.ipAddress, ld.forService, ld.referral
}

func (ld *hist) ID() int            { return ld.id }
func (ld *hist) UserID() int        { return ld.userID }
func (ld *hist) IPAddress() string  { return ld.ipAddress }
func (ld *hist) AcM() int           { return ld.acm }
func (ld *hist) Date() time.Time    { return ld.date }
func (ld *hist) ForService() string { return ld.forService }
func (ld *hist) Referral() string   { return ld.referral }
func (ld *hist) Successful() bool   { return ld.successful }

func TestNew(t *testing.T) {

	expLDets := initHistory()
	ld, err := history.New(expLDets.explodeParams())
	if err != nil {
		t.Fatalf("history.New(): %s", err)
	}

	compareHistory(ld, expLDets, t)
}

func initHistory() hist {
	return hist{
		userID:     userID,
		acm:        acm,
		successful: true,
		ipAddress:  ipAddress,
		date:       time.Now(),
		forService: refSrv,
		referral:   referall,
	}
}

func compareHistory(act *history.History, exp hist, t *testing.T) {

	if act == nil {
		t.Errorf("got nil, expected %v", exp)
		return
	}

	if act.UserID() != exp.userID {
		t.Errorf("Expected UserID %d but got %d", exp.UserID(), act.UserID())
	}
	if act.AccessMethod() != exp.AcM() {
		t.Errorf("Expected AccessMethod %d but got %d", exp.AcM(), act.AccessMethod())
	}
	if act.IPAddress() != exp.IPAddress() {
		t.Errorf("Expected IPAddress %s but got %s", exp.IPAddress(), act.IPAddress())
	}
	if act.Successful() != exp.Successful() {
		t.Errorf("Expected successful %v but got %v", exp.Successful(), act.Successful())
	}
	expDate := exp.Date().Truncate(1 * time.Second)
	actDate := exp.Date().Truncate(1 * time.Second)
	if !expDate.Equal(actDate) {
		t.Errorf("\nExp Date:\t%s\nGot:\t\t%s", expDate, actDate)
	}
	if act.ForService() != exp.ForService() {
		t.Errorf("Expected ForService %s but got %s", exp.ForService(), act.ForService())
	}
	if act.Referral() != exp.Referral() {
		t.Errorf("Expected Referral %s but got %s", exp.Referral(), act.Referral())
	}
}
