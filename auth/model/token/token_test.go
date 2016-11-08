package token_test

import (
	"testing"
	"time"

	"github.com/tomogoma/authms/auth/model/token"
)

const (
	userID       = 56789034561382845
	devID        = "93c66be8-f3df-460f-8c76-23497f22a267"
	shortExpTime = 8 * time.Hour
	medExpTime   = 720 * time.Hour
	longExpTime  = 4320 * time.Hour
)

type Token struct {
	id         int
	userID     int
	devID      string
	token      string
	issued     time.Time
	expAdd     time.Duration
	expiryType token.ExpiryType
}

func (t Token) ID() int           { return t.id }
func (t Token) UserID() int       { return t.userID }
func (t Token) DevID() string     { return t.devID }
func (t Token) Token() string     { return t.token }
func (t Token) Issued() time.Time { return t.issued }
func (t Token) Expiry() time.Time { return time.Now().Add(t.expAdd) }

func (t Token) explodeParams() (int, string, token.ExpiryType) {
	return t.userID, t.devID, t.expiryType
}

var expToken = Token{
	userID:     userID,
	devID:      devID,
	expAdd:     shortExpTime,
	expiryType: token.ShortExpType,
}

func TestNew_medDuration(t *testing.T) {

	act, err := token.New(userID, devID, token.MedExpType)
	if err != nil {
		t.Fatalf("token.New(): %s", err)
	}

	expToken.expAdd = medExpTime
	compareToken(act, expToken, t)
}

func TestNew_longDuration(t *testing.T) {

	act, err := token.New(userID, devID, token.LongExpType)
	if err != nil {
		t.Fatalf("token.New(): %s", err)
	}
	expToken.expAdd = longExpTime
	compareToken(act, expToken, t)
}

func TestNew_shortDuration(t *testing.T) {

	act, err := token.New(userID, devID, token.ShortExpType)
	if err != nil {
		t.Fatalf("token.New(): %s", err)
	}
	expToken.expAdd = shortExpTime
	compareToken(act, expToken, t)
}

func TestNew_BadExpiryType(t *testing.T) {

	act, err := token.New(userID, devID, 56321)
	if err != nil {
		t.Fatalf("token.New() with bad type: %s", err)
	}
	compareToken(act, expToken, t)
}

func TestNew_BadUserID(t *testing.T) {

	_, err := token.New(0, devID, token.ShortExpType)
	if err == nil || err != token.ErrorBadUserID {
		t.Fatalf("Expected error %s but got %s", token.ErrorBadUserID, err)
	}
}

func TestNew_EmptyDevID(t *testing.T) {

	_, err := token.New(userID, "", token.ShortExpType)
	if err == nil || err != token.ErrorEmptyDevID {
		t.Fatalf("Expected error %s but got %s", token.ErrorEmptyDevID, err)
	}
}

func compareToken(act token.Token, exp Token, t *testing.T) {

	if act.UserID() != exp.userID {
		t.Errorf("Expected UserID %s but got %s", exp.userID, act.UserID())
	}
	if act.DevID() != exp.devID {
		t.Errorf("Expected DevID %s but got %s", exp.devID, act.DevID())
	}
	if act.Token() == "" {
		t.Error("Expected non-empty Token")
	}
	expIssued := time.Now().Add(-1 * time.Minute)
	if act.Issued().Before(expIssued) {
		t.Errorf("Expected to be issued before %v but got %v", expIssued, act.Issued())
	}
	expExpiry := act.Issued().Add(exp.expAdd)
	if act.Expiry() != expExpiry {
		t.Errorf("Expected expiry %s but got %s", expExpiry, act.Expiry())
	}
}
