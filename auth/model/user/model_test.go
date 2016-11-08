package user_test

import (
	"database/sql"
	"testing"

	"github.com/tomogoma/authms/auth/model/helper"
	"github.com/tomogoma/authms/auth/model/testhelper"
	"github.com/tomogoma/authms/auth/model/user"
)

type testCase struct {
	desc    string
	appName string
	appUID  string
	call    func(app, appUID, pass string, hashFunc user.ValidatePassFunc) (user.User, error)
}

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
	i := save(expUser, m, t)
	if i < 1 {
		return
	}

	usr, err := m.Get(i)
	if err != nil {
		t.Fatalf("userModel.Get(): %s", err)
	}

	if usr == nil {
		t.Fatal("user was nil")
	}

	// TODO compare user values
}

func TestModel_Save_duplicateUser(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if i := save(expUser, m, t); i < 1 {
		return
	}

	usr, err := user.New(expUser.UName, "", "", "some-other-pass", nil, expUser.HashF)
	if err != nil {
		t.Fatalf("user.New(): %s", err)
	}
	_, err = m.Save(*usr)
	if err == nil || err != user.ErrorUserExists {
		t.Fatalf("Expected error %s but got %v", user.ErrorUserExists, err)
	}
}

func TestModel_GetBy(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if i := save(expUser, m, t); i < 1 {
		return
	}

	tcs := genericTestCases(m)

	for _, tc := range tcs {
		_, err := tc.call(tc.appName, tc.appUID,
			expUser.Password, expUser.ValHashFunc)
		if err != nil {
			t.Errorf("%s: Expected nil error but got %v",
				tc.desc, err)
		}
	}
}

func TestModel_GetBy_noUsers(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	tcs := genericTestCases(m)

	for _, tc := range tcs {
		_, err := tc.call(tc.appName, "some-username",
			expUser.Password, expUser.ValHashFunc)
		if err == nil || err != user.ErrorPasswordMismatch {
			t.Errorf("%s: Expected error %s but got %v",
				tc.desc, user.ErrorPasswordMismatch, err)
		}
	}
}

func TestModel_GetBy_identifierNotInDB(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if i := save(expUser, m, t); i < 1 {
		return
	}

	tcs := genericTestCases(m)

	for _, tc := range tcs {
		_, err := tc.call(tc.appName, "some-username",
			expUser.Password, expUser.ValHashFunc)
		if err == nil || err != user.ErrorPasswordMismatch {
			t.Errorf("%s: Expected error %s but got %v",
				tc.desc, user.ErrorPasswordMismatch, err)
		}
	}
}

func TestModel_GetBy_emptyIdentifier(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	if i := save(expUser, m, t); i < 1 {
		return
	}

	tcs := genericTestCases(m)

	for _, tc := range tcs {
		_, err := tc.call(tc.appName, "",
			expUser.Password, expUser.ValHashFunc)
		if err == nil || err != user.ErrorPasswordMismatch {
			t.Errorf("%s: Expected error %s but got %v",
				tc.desc, user.ErrorPasswordMismatch, err)
		}
	}
}

func TestModel_Get(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	uid := save(expUser, m, t)

	usr, err := m.Get(uid)
	if err != nil {
		t.Fatalf("userModel.Get(): %s", err)
	}

	if usr == nil {
		t.Fatal("got nil user")
	}

	if usr.ID() != uid {
		t.Errorf("Expected id %d got %d", uid, usr.ID())
	}
}

func TestModel_Get_noResults(t *testing.T) {

	m := newUserModel(t)
	defer testhelper.TearDown(db, t)

	_, err := m.Get(4567)
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

func newUserModel(t *testing.T) *user.Model {

	db = testhelper.SQLDB(t)
	m, err := user.NewModel(db)
	if err != nil {
		testhelper.TearDown(db, t)
		t.Fatalf("user.NewModel(): %s", err)
	}
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

func genericTestCases(m *user.Model) []testCase {
	return []testCase{
		{
			desc:   "GetByUsername",
			appUID: expUser.UserName(),
			call: func(app, appUID, pass string, hashFunc user.ValidatePassFunc) (user.User, error) {
				return m.GetByUserName(appUID, pass, hashFunc)
			},
		},
		{
			desc:   "GetByPhone",
			appUID: expUser.Phone().Value(),
			call: func(app, appUID, pass string, hashFunc user.ValidatePassFunc) (user.User, error) {
				return m.GetByPhone(appUID, pass, hashFunc)
			},
		},
		{
			desc:   "GetByEmail",
			appUID: expUser.Email().Value(),
			call: func(app, appUID, pass string, hashFunc user.ValidatePassFunc) (user.User, error) {
				return m.GetByEmail(appUID, pass, hashFunc)
			},
		},
		{
			desc:    "GetByAppUserID",
			appName: expUser.App().Name(),
			appUID:  expUser.App().UserID(),
			call: func(app, appUID, pass string, hashFunc user.ValidatePassFunc) (user.User, error) {
				return m.GetByAppUserID(app, appUID, pass, hashFunc)
			},
		},
	}
}
