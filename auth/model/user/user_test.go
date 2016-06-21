package user_test

import (
	"testing"

	"errors"

	"bitbucket.org/tomogoma/auth-ms/auth/model/user"
	"golang.org/x/crypto/bcrypt"
)

const (
	uname = "johndoe"
	pass  = "zCJ\"~6x4"
	fname = "John"
	mname = "Nyawira"
	lname = "Doe"
)

type User struct {
	UserName string
	Password string

	FirstName  string
	MiddleName string
	LastName   string
	hashF      user.HashFunc
}

func (u User) explodeParams() (string, string, string, string, string, user.HashFunc) {
	return u.UserName, u.FirstName, u.MiddleName, u.LastName, u.Password, u.hashF
}

var errorHashing = errors.New("some hashing error")
var expUser = User{
	UserName:   uname,
	Password:   pass,
	FirstName:  fname,
	MiddleName: mname,
	LastName:   lname,
	hashF:      hashF,
}

func hashF(p string) ([]byte, error) {
	return []byte{0, 1, 2, 3, 4, 5}, nil
}

func anotherHashF(p string) ([]byte, error) {
	return []byte{6, 7, 8, 9, 10}, nil
}

func errorHashF(p string) ([]byte, error) {
	return nil, errorHashing
}

func TestNew(t *testing.T) {

	act, err := user.New(expUser.explodeParams())
	if err != nil {
		t.Fatalf("user.New(): %s", err)
	}

	compareUsersShallow(act, expUser, t)
}

func TestNew_emptyUserName(t *testing.T) {

	_, err := user.New("", fname, mname, lname, pass, hashF)
	if err == nil || err != user.ErrorEmptyUserName {
		t.Fatalf("Expected error %s but got %s", user.ErrorEmptyUserName, err)
	}
}

func TestNew_emptyPassword(t *testing.T) {

	_, err := user.New(uname, fname, mname, lname, "", hashF)
	if err == nil || err != user.ErrorEmptyPassword {
		t.Fatalf("Expected error %s but got %s", user.ErrorEmptyPassword, err)
	}
}

func TestNew_NilHashFunc(t *testing.T) {

	_, err := user.New(uname, fname, mname, lname, pass, nil)
	if err == nil || err != user.ErrorNilHashFunc {
		t.Fatalf("Expected error %s but got %s", user.ErrorNilHashFunc, err)
	}
}

func TestNew_HashFuncReportError(t *testing.T) {

	_, err := user.New(uname, fname, mname, lname, pass, errorHashF)
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

func compareUsersShallow(act user.User, exp User, t *testing.T) {

	if act.UserName() != exp.UserName {
		t.Errorf("Expected UserName %s but got %s", exp.UserName, act.UserName())
	}
	if act.FirstName() != exp.FirstName {
		t.Errorf("Expected FirstName %s but got %s", exp.FirstName, act.FirstName())
	}
	if act.MiddleName() != exp.MiddleName {
		t.Errorf("Expected MiddleName %s but got %s", exp.MiddleName, act.MiddleName())
	}
	if act.LastName() != exp.LastName {
		t.Errorf("Expected LastName %s but got %s", exp.LastName, act.LastName())
	}
}
