package history_test

import (
	"database/sql"
	"testing"

	"time"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/history"
	"bitbucket.org/tomogoma/auth-ms/auth/model/testhelper"
)

var db *sql.DB

func TestNewModel(t *testing.T) {
	newModel(t)
	defer testhelper.TearDown(db, t)
}

func TestNewModel_nilDB(t *testing.T) {

	_, err := history.NewModel(nil)
	if err == nil || err != helper.ErrorNilDB {
		t.Fatalf("Expected error %s but got %v", helper.ErrorNilDB, err)
	}
}

func TestModel_Save_invalidHistory(t *testing.T) {

	m := newModel(t)
	defer testhelper.TearDown(db, t)
	_, err := m.Save(history.History{})
	if err == nil {
		t.Fatalf("Expected an error but got nil")
	}
}

func TestModel_Save_n_Get(t *testing.T) {

	m := newModel(t)
	defer testhelper.TearDown(db, t)

	det1 := hist{userID: 1956, acm: history.LoginAccess}
	save(det1, m, t)
	det2 := hist{userID: userID, acm: history.TokenValidationAccess}
	save(det2, m, t)
	exp := initHistory()
	save(exp, m, t)

	acts, err := m.Get(exp.UserID(), 0, 10)
	if err != nil {
		t.Fatalf("historyModel.Get(): %s", err)
	}

	if len(acts) != 2 {
		t.Fatalf("Expected 2 history details but got %d", len(acts))
	}

	compareHistory(acts[0], exp, t)
	compareHistory(acts[1], det2, t)
}

func TestModel_Get_fetchSpecificType(t *testing.T) {

	m := newModel(t)
	defer testhelper.TearDown(db, t)

	det1 := hist{userID: 1956, acm: history.LoginAccess}
	save(det1, m, t)
	det2 := hist{userID: userID, acm: history.TokenValidationAccess}
	save(det2, m, t)
	exp := initHistory()
	save(exp, m, t)

	acts, err := m.Get(exp.UserID(), 0, 10, history.LoginAccess)
	if err != nil {
		t.Fatalf("historyModel.Get(): %s", err)
	}

	if len(acts) != 1 {
		t.Errorf("Expected 1 history item but got %d", len(acts))
	}

	for _, act := range acts {
		compareHistory(act, exp, t)
	}
}

func TestModel_Get_fetchMultipleTypes(t *testing.T) {

	m := newModel(t)
	defer testhelper.TearDown(db, t)

	det1 := hist{userID: userID, acm: history.LoginAccess, date: time.Now()}
	save(det1, m, t)
	det2 := hist{userID: userID, acm: history.TokenValidationAccess, date: time.Now()}
	save(det2, m, t)
	exp := initHistory()
	save(exp, m, t)

	acts, err := m.Get(exp.UserID(), 0, 10, history.LoginAccess, history.TokenValidationAccess)
	if err != nil {
		t.Fatalf("historyModel.Get(): %s", err)
	}

	if len(acts) != 3 {
		t.Fatalf("Expected 3 history details but got %d", len(acts))
	}

	compareHistory(acts[0], exp, t)
	compareHistory(acts[1], det2, t)
	compareHistory(acts[2], det1, t)
}

func TestModel_Get_noResults(t *testing.T) {

	m := newModel(t)
	defer testhelper.TearDown(db, t)

	acts, err := m.Get(523347789, 0, 10)
	if err != nil {
		t.Fatalf("historyModel.Get(): %s", err)
	}

	if len(acts) != 0 {
		t.Errorf("Expected 0 history items but got %d", len(acts))
	}
}

func newModel(t *testing.T) *history.Model {

	db = testhelper.InstantiateDB(t)
	m, err := history.NewModel(db)
	if err != nil {
		t.Fatalf("history.NewModel(): %s", err)
	}
	testhelper.SetUp(m, db, t)

	return m
}

func save(h hist, m *history.Model, t *testing.T) {

	ld, err := history.New(h.explodeParams())
	if err != nil {
		t.Fatalf("history.New(): %s", err)
	}

	i, err := m.Save(*ld)
	if err != nil {
		t.Fatalf("historyModel.Save(): %s", err)
	}

	if i < 1 {
		t.Errorf("Expected id > 1 but got %d", i)
	}
}
