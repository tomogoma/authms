package login_test

import (
	"testing"
	"time"

	"bitbucket.org/tomogoma/auth-ms/auth/model/details/login"
)

const (
	userID    = 56789034561382845
	ipAddress = "127.0.0.1"
	refSrv    = "SOME-SERV-ID"
	referall  = "SOME-OTHER-SERV-ID"
)

type loginDetails struct {
	id         int
	userID     int
	ipAddress  string
	date       time.Time
	forService string
	referral   string
}

func (ld *loginDetails) explodeParams() (int, time.Time, string, string, string) {
	return ld.userID, ld.date, ld.ipAddress, ld.forService, ld.referral
}

func (ld *loginDetails) ID() int            { return ld.id }
func (ld *loginDetails) UserID() int        { return ld.userID }
func (ld *loginDetails) IPAddress() string  { return ld.ipAddress }
func (ld *loginDetails) Date() time.Time    { return ld.date }
func (ld *loginDetails) ForService() string { return ld.forService }
func (ld *loginDetails) Referral() string   { return ld.referral }

func TestNew(t *testing.T) {

	expLDets := initLoginDets()
	ld, err := login.New(expLDets.explodeParams())
	if err != nil {
		t.Fatalf("login.New(): %s", err)
	}

	compareLoginDets(ld, expLDets, t)
}

func initLoginDets() loginDetails {
	return loginDetails{
		userID:     userID,
		ipAddress:  ipAddress,
		date:       time.Now(),
		forService: refSrv,
		referral:   referall,
	}
}

func compareLoginDets(act *login.LoginDetails, exp loginDetails, t *testing.T) {

	if act == nil {
		t.Errorf("got nil, expected %v", exp)
		return
	}

	if act.UserID() != exp.userID {
		t.Errorf("Expected UserID % but got %s", exp.UserID(), act.UserID())
	}
	if act.IPAddress() != exp.IPAddress() {
		t.Errorf("Expected IPAddress % but got %s", exp.IPAddress(), act.IPAddress())
	}
	expDate := exp.Date().Truncate(1 * time.Second)
	actDate := exp.Date().Truncate(1 * time.Second)
	if !expDate.Equal(actDate) {
		t.Errorf("\nExp Date:\t%s\nGot:\t\t%s", expDate, actDate)
	}
	if act.ForService() != exp.ForService() {
		t.Errorf("Expected ForService % but got %s", exp.ForService(), act.ForService())
	}
	if act.Referral() != exp.Referral() {
		t.Errorf("Expected Referral % but got %s", exp.Referral(), act.Referral())
	}
}
