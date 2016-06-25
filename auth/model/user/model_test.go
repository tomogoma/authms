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

	if i := save(expUser, m, t); i < 1 {
		return
	}

	usr, err := m.Get(expUser.UserName(), expUser.Password, expUser.ValHashFunc)
	if err != nil {
		t.Fatalf("userModel.Get(): %s", err)
	}

	compareUsersShallow(usr, expUser, t)
}

func TestModel_Get_PassMismatch(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if i := save(expUser, m, t); i < 1 {
		return
	}

	u, err := m.Get(expUser.UserName(), expUser.Password, invalidHashF)
	if err == nil || err != user.ErrorPasswordMismatch {
		t.Fatalf("expected error %s but got %s", errorHashing, err)
	}

	compareUsersShallow(u, expUser, t)
}

func TestModel_Get_noUsers(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	usr, err := m.Get(expUser.UserName(), expUser.Password, expUser.ValHashFunc)
	if err == nil || err != user.ErrorPasswordMismatch {
		t.Fatalf("Expected error %s but got %v", user.ErrorPasswordMismatch, err)
	}

	if usr != nil {
		t.Errorf("Expected nil user but got %v", usr)
	}
}

func TestModel_Get_userNameNotInDB(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if i := save(expUser, m, t); i < 1 {
		return
	}

	usr, err := m.Get("someUserName", expUser.Password, expUser.ValHashFunc)
	if err == nil || err != user.ErrorPasswordMismatch {
		t.Fatalf("Expected error %s but got %v", user.ErrorPasswordMismatch, err)
	}

	if usr != nil {
		t.Errorf("Expected nil user but got %v", usr)
	}
}

func TestModel_Get_emptyUserName(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if i := save(expUser, m, t); i < 1 {
		return
	}

	usr, err := m.Get("", expUser.Password, expUser.ValHashFunc)
	if err == nil || err != user.ErrorPasswordMismatch {
		t.Fatalf("Expected error %s but got %v", user.ErrorPasswordMismatch, err)
	}

	if usr != nil {
		t.Errorf("Expected nil user but got %v", usr)
	}
}

func TestModel_GetByID(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	uid := save(expUser, m, t)

	usr, err := m.GetByID(uid)
	if err != nil {
		t.Fatalf("userModel.Get(): %s", err)
	}

	compareUsersShallow(usr, expUser, t)
}

func TestModel_GetByID_noResults(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	usr, err := m.GetByID(4567)
	if err != nil {
		t.Fatalf("userModel.Get(): %s", err)
	}

	if usr != nil {
		t.Fatalf("Expected nil user but got %v", usr)
	}
}

func newUserModel(t *testing.T) *user.Model {

	db = testhelper.InstantiateDB(t)
	m, err := user.NewModel(db)
	if err != nil {
		testhelper.TearDown(db, t)
		t.Fatalf("user.NewModel(): %s", err)
	}
	testhelper.SetUp(m, db, t)
	return m
}

func save(expU testhelper.User, m *user.Model, t *testing.T) int {

	u, err := user.New(expU.ExplodeParams())
	if err != nil {
		t.Fatalf("user.New(): %s", err)
		return 0
	}

	us, err := m.Save(*u)
	if err != nil {
		t.Fatalf("userModel.Save(): %s", err)
		return 0
	}

	if us.ID() < 1 {
		t.Errorf("Expected id > 1 got %d", us.ID())
		return 0
	}
	return us.ID()
}
