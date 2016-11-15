package user_test

import (
	"testing"

	"errors"

	"github.com/tomogoma/authms/auth/model/testhelper"
	"github.com/tomogoma/authms/auth/model/user"
	"github.com/tomogoma/authms/auth/password"
	"golang.org/x/crypto/bcrypt"
)

const (
	uname     = "johndoe"
	email     = "johndoe@email.com"
	phone     = "+254712345678"
	appUserID = "bb7d8e35-1d05-4a76-9c82-dbb261a7b6b8"
	appName   = "facebook"
	pass      = "zCJ\"~6x4"
)

var errorHashing = errors.New("some hashing error")
var expUser = testhelper.User{
	UName:       uname,
	Password:    pass,
	EmailAddr:   &testhelper.Value{Val: email},
	PhoneNo:     &testhelper.Value{Val: phone},
	AppDet:      &testhelper.App{AppUID: appUserID, AppName: appName},
	HashF:       testhelper.HashF,
	ValHashFunc: testhelper.ValHashFunc,
}

func errorHashF(p string) ([]byte, error) {
	return nil, errorHashing
}

func TestNew(t *testing.T) {

	act, err := user.New(expUser.ExplodeParams())
	if err != nil {
		t.Fatalf("user.New(): %s", err)
	}

	compareUsersShallow(act, expUser, t)
}

func passwordGenerator(t *testing.T) *password.Generator {
	g, err := password.NewGenerator(password.AllChars)
	if err != nil {
		t.Fatalf("password.NewGenerator(): %s", err)
	}
	return g
}

func TestNew_noIdentifier(t *testing.T) {
	_, err := user.New("", "", "", "", nil, passwordGenerator(t), testhelper.HashF)
	if err == nil || err != user.ErrorEmptyIdentifier {
		t.Fatalf("Expected error %s but got %s", user.ErrorEmptyIdentifier, err)
	}
}

func TestNew_emptyPassword(t *testing.T) {
	_, err := user.New(uname, phone, email, "",
		&testhelper.App{AppUID: appUserID, AppName: appName},
		passwordGenerator(t), testhelper.HashF)
	if err == nil || err != user.ErrorEmptyPassword {
		t.Fatalf("Expected error %s but got %s", user.ErrorEmptyPassword, err)
	}
}

func TestNew_nilGenerator(t *testing.T) {
	_, err := user.New("", "", "", "", &testhelper.App{AppUID: appUserID, AppName: appName},
		nil, testhelper.HashF)
	if err == nil || err != user.ErrorNilPasswordGenerator {
		t.Fatalf("Expected error %s but got %s", user.ErrorNilPasswordGenerator, err)
	}
}

func TestNew_NilHashFunc(t *testing.T) {

	_, err := user.New(uname, phone, email, "some-password",
		&testhelper.App{AppUID: appUserID, AppName: appName},
		passwordGenerator(t), nil)
	if err == nil || err != user.ErrorNilHashFunc {
		t.Fatalf("Expected error %s but got %s", user.ErrorNilHashFunc, err)
	}
}

func TestNew_HashFuncReportError(t *testing.T) {

	_, err := user.New(uname, phone, email, "some-password",
		&testhelper.App{AppUID: appUserID, AppName: appName},
		passwordGenerator(t), errorHashF)
	if err == nil || err != errorHashing {
		t.Fatalf("Expected error %s but got %s", errorHashing, err)
	}
}

func TestHash(t *testing.T) {

	passHB, err := user.Hash(pass)
	if err != nil {
		t.Fatalf("user.Hash(): %s", err)
	}

	passB := []byte(pass)
	err = bcrypt.CompareHashAndPassword(passHB, passB)
	if err != nil {
		t.Fatalf("Password produced does not match bcrypt password: %s", err)
	}
}

func compareUsersShallow(act user.User, exp testhelper.User, t *testing.T) {

	if act.UserName() != exp.UserName() {
		t.Errorf("Expected UserName %+v but got %+v", exp.UserName(), act.UserName())
	}
	//defer func() { recover() }()
	actApp := act.App()
	expApp := exp.App()
	if expApp != nil && actApp != nil {
		if actApp.Name() != expApp.Name() {
			t.Errorf("Expected app name %+v but got %+v",
				expApp.Name(), actApp.Name())
		}
		if actApp.UserID() != expApp.UserID() {
			t.Errorf("Expected app user ID %+v but got %+v",
				expApp.UserID(), actApp.UserID())
		}
		if actApp.Validated() != expApp.Validated() {
			t.Errorf("Expected app validated %+v but got %+v",
				expApp.Validated(), actApp.Validated())
		}
	}
	compareValue(act.Email(), exp.Email(), "email value", t)
	compareValue(act.Phone(), exp.Phone(), "email value", t)
}

func compareValue(exp user.Valuer, act user.Valuer, desc string, t *testing.T) {
	if exp.Value() != act.Value() {
		t.Errorf("Expected %s %s but got %s", desc, exp.Value(), act.Value())
	}
}
