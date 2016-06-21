package login_test

import (
	"database/sql"
	"testing"

	"bitbucket.org/tomogoma/auth-ms/auth/model/details/login"
	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/testhelper"
)

var db *sql.DB

func TestNewModel(t *testing.T) {
	newModel(t)
	defer testhelper.TearDown(db, t)
}

func TestNewModel_nilDB(t *testing.T) {

	_, err := login.NewModel(nil)
	if err == nil || err != helper.ErrorNilDB {
		t.Fatalf("Expected error %s but got %v", helper.ErrorNilDB, err)
	}
}

func TestModel_Save_Get(t *testing.T) {

	m := newModel(t)
	defer testhelper.TearDown(db, t)

	exp := initLoginDets()
	ld, err := login.New(exp.explodeParams())
	if err != nil {
		t.Fatalf("login.New(): %s", err)
	}

	i, err := m.Save(*ld)
	if err != nil {
		t.Fatalf("loginModel.Save(): %s", err)
	}

	if i < 1 {
		t.Errorf("Expected id > 1 but got %d", i)
	}

	acts, err := m.Get(exp.UserID(), 0, 10)
	if err != nil {
		t.Fatalf("loginModel.Get(): %s", err)
	}

	if len(acts) != 1 {
		t.Errorf("Expected 1 login detail but got %d", len(acts))
	}

	for _, act := range acts {
		compareLoginDets(act, exp, t)
	}
}

func TestModel_Get_noResults(t *testing.T) {

	m := newModel(t)
	defer testhelper.TearDown(db, t)

	acts, err := m.Get(523347789, 0, 10)
	if err != nil {
		t.Fatalf("loginModel.Get(): %s", err)
	}

	if len(acts) != 0 {
		t.Errorf("Expected 0 login details but got %d", len(acts))
	}
}

func newModel(t *testing.T) *login.Model {

	db = testhelper.InstantiateDB(t)
	m, err := login.NewModel(db)
	if err != nil {
		t.Fatalf("login.NewModel(): %s", err)
	}
	testhelper.SetUp(m, db, t)

	return m
}
