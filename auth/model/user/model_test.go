package user_test

import (
	"database/sql"
	"testing"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/testhelper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/user"
	_ "github.com/lib/pq"
)

var db *sql.DB

func TestNewModel(t *testing.T) {

	newModel(t)
	defer testhelper.TearDown(db, t)
}

func TestNewModel_nilDB(t *testing.T) {

	_, err := user.NewModel(nil)
	if err == nil || err != helper.ErrorNilDB {
		t.Fatalf("Expected error %s but got %s", helper.ErrorNilDB, err)
	}
}

func TestModel_Save_n_Get(t *testing.T) {

	m := newModel(t)
	defer testhelper.TearDown(db, t)

	save(expUser, m, t)
	usr, err := m.Get(expUser.UserName, expUser.Password, expUser.hashF)
	if err != nil {
		t.Fatalf("model.Get(): %s", err)
	}

	compareUsersShallow(usr, expUser, t)
}

func TestModel_Get_hashError(t *testing.T) {

	m := newModel(t)
	defer testhelper.TearDown(db, t)

	save(expUser, m, t)
	_, err := m.Get(expUser.UserName, expUser.Password, errorHashF)
	if err == nil || err != errorHashing {
		t.Fatalf("expected error %s but got %s", errorHashing, err)
	}
}

func newModel(t *testing.T) *user.Model {

	db = testhelper.InstantiateDB(t)
	m, err := user.NewModel(db)
	if err != nil {
		t.Fatalf("user.NewModel(): %s", err)
	}
	testhelper.SetUp(m, db, t)
	return m
}

func save(expU User, m *user.Model, t *testing.T) {

	u, err := user.New(expU.explodeParams())
	if err != nil {
		t.Fatalf("user.New(): %s", err)
	}

	i, err := m.Save(*u)
	if err != nil {
		t.Fatalf("model.Save(): %s", err)
	}

	if i < 1 {
		t.Errorf("Expected id > 1 got %d", i)
	}
}
