package user_test

import (
	"database/sql"
	"testing"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/testhelper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/user"
)

var db *sql.DB

func TestNewModel(t *testing.T) {

	newUserModel(t)
	defer testhelper.TearDown(db, t)
}

func TestNewModel_nilDB(t *testing.T) {

	_, err := user.NewModel(nil)
	if err == nil || err != helper.ErrorNilDB {
		t.Fatalf("Expected error %s but got %s", helper.ErrorNilDB, err)
	}
}

func TestModel_Save_n_Get(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if !save(expUser, m, t) {
		return
	}

	usr, err := m.Get(expUser.UserName, expUser.Password, expUser.hashF)
	if err != nil {
		t.Fatalf("userModel.Get(): %s", err)
	}

	compareUsersShallow(usr, expUser, t)
}

func TestModel_Get_hashError(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if !save(expUser, m, t) {
		return
	}

	_, err := m.Get(expUser.UserName, expUser.Password, errorHashF)
	if err == nil || err != errorHashing {
		t.Fatalf("expected error %s but got %s", errorHashing, err)
	}
}

func TestModel_Get_noUsers(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	usr, err := m.Get(expUser.UserName, expUser.Password, expUser.hashF)
	if err != nil {
		t.Fatalf("userModel.Get(): %s", err)
	}

	if usr != nil {
		t.Fatalf("Expected nil user but got %v", usr)
	}
}

func TestModel_Get_userNameNotInDB(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if !save(expUser, m, t) {
		return
	}

	usr, err := m.Get("someUserName", expUser.Password, expUser.hashF)
	if err != nil {
		t.Fatalf("userModel.Get(): %s", err)
	}

	if usr != nil {
		t.Fatalf("Expected nil user but got %v", usr)
	}
}

func TestModel_Get_passNotInDB(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if !save(expUser, m, t) {
		return
	}

	usr, err := m.Get(expUser.UserName, "some other password", anotherHashF)
	if err != nil {
		t.Fatalf("userModel.Get(): %s", err)
	}

	if usr != nil {
		t.Fatalf("Expected nil user but got %v", usr)
	}
}

func newUserModel(t *testing.T) *user.Model {

	db = testhelper.InstantiateDB(t)
	return testhelper.NewUserModel(db, t)
}

func save(expU User, m *user.Model, t *testing.T) bool {

	u, err := user.New(expU.explodeParams())
	if err != nil {
		t.Fatalf("user.New(): %s", err)
		return false
	}

	i, err := m.Save(*u)
	if err != nil {
		t.Fatalf("userModel.Save(): %s", err)
		return false
	}

	if i < 1 {
		t.Errorf("Expected id > 1 got %d", i)
		return false
	}
	return true
}
