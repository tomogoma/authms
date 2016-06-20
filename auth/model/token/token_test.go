package token_test

import (
	"testing"
	"time"

	"bitbucket.org/tomogoma/auth-ms/auth/model/token"
)

const (
	userID = 56789034561382845
	devID  = "93c66be8-f3df-460f-8c76-23497f22a267"
)

type Token struct {
	ID     int
	UserID int
	DevID  string
	Token  string
	Issued time.Time
	Expiry time.Time
}

var expToken = Token{}

func TestNew_medDuration(t *testing.T) {

	act, err := token.New(userID, devID, token.MedExpType)
	if err != nil {
		t.Fatalf("token.New(): %s", err)
	}

	compareToken(act, expToken, t)
}

func TestNew_longDuration(t *testing.T) {

	act, err := token.New(userID, devID, token.LongExpType)
	if err != nil {
		t.Fatalf("token.New(): %s", err)
	}
	compareToken(act, expToken, t)
}

func TestNew_shortDuration(t *testing.T) {

	act, err := token.New(userID, devID, token.ShortExpType)
	if err != nil {
		t.Fatalf("token.New(): %s", err)
	}
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
	t.Errorf("Not yet implemented")
}
