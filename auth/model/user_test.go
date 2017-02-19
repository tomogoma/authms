package model_test

import (
	"github.com/tomogoma/authms/proto/authms"
	"reflect"
	"database/sql"
	"testing"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/authms/auth/model"
)

type UpdateTestCase struct {
	Desc string
	User *authms.User
}

func TestModel_SaveUser(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	m := newModel(t)
	usr := &authms.User{
		OAuth: &authms.OAuth{
			AppName: "test-app",
			AppUserID: "test-user-id",
			AppToken: "test-app.test-user-id.test-token",
		},
	}
	if err := m.SaveUser(usr); err != nil {
		t.Fatalf("model.Save(): %s", err)
	}
	dbUsr := fetchUser(usr.ID, t)
	assertUsersEqual(dbUsr, usr, t)
}

func TestModel_UpdateAppUserID(t *testing.T) {
	bareBoneUser := &authms.User{
		UserName: "test-username",
		Password: "test-password",
	}
	cmpltUsr := completeUser()
	appName := cmpltUsr.OAuth.AppName
	appToken := cmpltUsr.OAuth.AppToken
	tcs := []UpdateTestCase{
		{Desc: "Is to update", User: cmpltUsr},
		{Desc: "Is to insert", User: bareBoneUser},
	}
	for _, tc := range tcs {
		func() {
			expUsr := tc.User
			setUp(t)
			defer tearDown(t)
			m := newModel(t)
			insertUser(tc.User, t)
			if expUsr.OAuth == nil {
				expUsr.OAuth = &authms.OAuth{
					AppName: appName,
					AppToken: appToken,
				}
			}
			expUsr.OAuth.AppUserID = "test-user-id-updated"
			tkn, err := tg.Generate(int(expUsr.ID), "test-dev", token.MedExpType)
			if err != nil {
				t.Errorf("%s - token.Generate(): %s", tc.Desc, err)
				return
			}
			err = m.UpdateAppUserID(tkn.Token(), expUsr.OAuth)
			if err != nil {
				t.Errorf("%s - model.UpdateAppUserID(): %s", tc.Desc, err)
				return
			}
			dbUsr := fetchUser(expUsr.ID, t)
			assertUsersEqual(dbUsr, expUsr, t)
		}()
	}
}

func TestModel_UpdatePhone(t *testing.T) {
	bareBoneUser := &authms.User{
		UserName: "test-username",
		Password: "test-password",
	}
	cmpltUsr := completeUser()
	tcs := []UpdateTestCase{
		{Desc: "Is to update", User: cmpltUsr},
		{Desc: "Is to insert", User: bareBoneUser},
	}
	for _, tc := range tcs {
		func() {
			expUsr := tc.User
			setUp(t)
			defer tearDown(t)
			m := newModel(t)
			insertUser(tc.User, t)
			if expUsr.Phone == nil {
				expUsr.Phone = &authms.Value{}
			}
			expUsr.Phone.Value = "+254098765432"
			tkn, err := tg.Generate(int(expUsr.ID), "test-dev", token.MedExpType)
			if err != nil {
				t.Errorf("%s - token.Generate(): %s", tc.Desc, err)
				return
			}
			err = m.UpdatePhone(tkn.Token(), expUsr.Phone.Value)
			if err != nil {
				t.Errorf("%s - model.UpdateAppUserID(): %s", tc.Desc, err)
				return
			}
			dbUsr := fetchUser(expUsr.ID, t)
			assertUsersEqual(dbUsr, expUsr, t)
		}()
	}
}

func TestModel_UpdateEmail(t *testing.T) {
	bareBoneUser := &authms.User{
		UserName: "test-username",
		Password: "test-password",
	}
	cmpltUsr := completeUser()
	tcs := []UpdateTestCase{
		{Desc: "Is to update", User: cmpltUsr},
		{Desc: "Is to insert", User: bareBoneUser},
	}
	for _, tc := range tcs {
		func() {
			expUsr := tc.User
			setUp(t)
			defer tearDown(t)
			m := newModel(t)
			insertUser(tc.User, t)
			if expUsr.Email == nil {
				expUsr.Email = &authms.Value{}
			}
			expUsr.Email.Value = "test.update@email.com"
			tkn, err := tg.Generate(int(expUsr.ID), "test-dev", token.MedExpType)
			if err != nil {
				t.Errorf("%s - token.Generate(): %s", tc.Desc, err)
				return
			}
			err = m.UpdateEmail(tkn.Token(), expUsr.Email.Value)
			if err != nil {
				t.Errorf("%s - model.UpdateAppUserID(): %s", tc.Desc, err)
				return
			}
			dbUsr := fetchUser(expUsr.ID, t)
			assertUsersEqual(dbUsr, expUsr, t)
		}()
	}
}

func TestModel_UpdateUserName(t *testing.T) {
	bareBoneUser := &authms.User{
		Phone: &authms.Value{Value: "+254012345678"},
		Password: "test-password",
	}
	cmpltUsr := completeUser()
	tcs := []UpdateTestCase{
		{Desc: "Is to update", User: cmpltUsr},
		{Desc: "Is to insert", User: bareBoneUser},
	}
	for _, tc := range tcs {
		func() {
			expUsr := tc.User
			setUp(t)
			defer tearDown(t)
			m := newModel(t)
			insertUser(tc.User, t)
			expUsr.UserName = "test-updated-username"
			tkn, err := tg.Generate(int(expUsr.ID), "test-dev", token.MedExpType)
			if err != nil {
				t.Errorf("%s - token.Generate(): %s", tc.Desc, err)
				return
			}
			err = m.UpdateUserName(tkn.Token(), expUsr.UserName)
			if err != nil {
				t.Errorf("%s - model.UpdateAppUserID(): %s", tc.Desc, err)
				return
			}
			dbUsr := fetchUser(expUsr.ID, t)
			assertUsersEqual(dbUsr, expUsr, t)
		}()
	}
}

func TestModel_UpdatePassword(t *testing.T) {
	type PasswordTestCase struct {
		Desc    string
		OldPass string
		expErr  error
	}
	expUsr := completeUser()
	tcs := []PasswordTestCase{
		{Desc: "Correct old password", OldPass: expUsr.Password,
			expErr: nil},
		{Desc: "Incorrect old password", OldPass: "some-invalid",
			expErr: model.ErrorPasswordMismatch},
	}
	newPass := "test-updated-password"
	for _, tc := range tcs {
		func() {
			setUp(t)
			defer tearDown(t)
			m := newModel(t)
			insertUser(expUsr, t)
			err := m.UpdatePassword(expUsr.ID, tc.OldPass, newPass)
			if err != tc.expErr {
				t.Errorf("%s - model.UpdatePassword() expected" +
					" error %v but got %v", tc.Desc, tc.expErr, err)
				return
			}
			if tc.expErr != nil {
				return
			}
			db := getDB(t)
			q := `SELECT password FROM users WHERE id=$1`
			var updatedPassHB []byte
			if err = db.QueryRow(q, expUsr.ID).Scan(&updatedPassHB); err != nil {
				t.Fatalf("%s - error retrieving password for" +
					" validation: %v", tc.Desc, err)
			}
			if !hasher.CompareHash(newPass, updatedPassHB) {
				t.Error("New password in db did not match " +
					"new password provided")
			}
		}()
	}
}

func TestModel_GetByAppUserID(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	m := newModel(t)
	expUsr := completeUser()
	insertUser(expUsr, t)
	actUser, err := m.GetByAppUserID(expUsr.OAuth.AppName,
		expUsr.OAuth.AppUserID, "")
	if err != nil {
		t.Fatalf("model.Get(): %s", err)
	}
	assertUsersEqual(actUser, expUsr, t)
}

func TestModel_GetByEmail(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	m := newModel(t)
	expUsr := completeUser()
	insertUser(expUsr, t)
	actUser, err := m.GetByEmail(expUsr.Email.Value, expUsr.Password)
	if err != nil {
		t.Fatalf("model.Get(): %s", err)
	}
	assertUsersEqual(actUser, expUsr, t)
}

func TestModel_GetByPhone(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	m := newModel(t)
	expUsr := completeUser()
	insertUser(expUsr, t)
	actUser, err := m.GetByPhone(expUsr.Phone.Value, expUsr.Password)
	if err != nil {
		t.Fatalf("model.Get(): %s", err)
	}
	assertUsersEqual(actUser, expUsr, t)
}

func TestModel_GetByUserName(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	m := newModel(t)
	expUsr := completeUser()
	insertUser(expUsr, t)
	actUser, err := m.GetByUserName(expUsr.UserName, expUsr.Password)
	if err != nil {
		t.Fatalf("model.Get(): %s", err)
	}
	assertUsersEqual(actUser, expUsr, t)
}

func completeUser() (*authms.User) {
	return &authms.User{
		UserName: "test-username",
		Email: &authms.Value{
			Value: "test@email.com",
			Verified: true,
		},
		Phone: &authms.Value{
			Value: "+254712345678",
			Verified: false,
		},
		OAuth: &authms.OAuth{
			AppName: "test-app",
			AppUserID: "test-user-id",
			AppToken: "test-app.test-user-id.test-token",
		},
		Password: "test-password",
	}
}

func fetchUser(userID int64, t *testing.T) *authms.User {
	db := getDB(t)
	query := `
	SELECT users.id, userNames.userName, phones.phone, phones.validated,
		  emails.email, emails.validated, appUserIDs.appName,
		  appUserIDs.appUserID
		FROM users
		LEFT JOIN userNames ON users.id=userNames.userID
		LEFT JOIN phones ON users.id=phones.userID
		LEFT JOIN emails ON users.id=emails.userID
		LEFT JOIN appUserIDs ON users.id=appUserIDs.userID
		WHERE users.id=$1`
	dbUsr := &authms.User{
		Email: &authms.Value{},
		Phone: &authms.Value{},
		OAuth: &authms.OAuth{},
	}
	var dbUserName, dbPhone, dbEmail, dbAppName, dbAppUsrID sql.NullString
	var dbPhoneValidated, dbEmailValidated sql.NullBool
	err := db.QueryRow(query, userID).Scan(&dbUsr.ID, &dbUserName, &dbPhone,
		&dbPhoneValidated, &dbEmail, &dbEmailValidated, &dbAppName,
		&dbAppUsrID)
	dbUsr.UserName = dbUserName.String
	dbUsr.Phone.Value = dbPhone.String
	dbUsr.Phone.Verified = dbPhoneValidated.Bool
	dbUsr.Email.Value = dbEmail.String
	dbUsr.Email.Verified = dbEmailValidated.Bool
	dbUsr.OAuth.AppName = dbAppName.String
	dbUsr.OAuth.AppUserID = dbAppUsrID.String
	if err != nil {
		t.Fatalf("Error checking db contents for validation: %s", err)
	}
	return dbUsr
}

func insertUser(u *authms.User, t *testing.T) {
	db := getDB(t)
	query := `INSERT INTO users (password, createDate)
			VALUES($1, CURRENT_TIMESTAMP())
			RETURNING id`
	pass := u.Password
	if pass == "" {
		pass = "some-random-password"
	}
	passHB, err := hasher.Hash(pass)
	if err != nil {
		t.Fatalf("Error setting up (hashing password): %s", err)
	}
	if err := db.QueryRow(query, passHB).
		Scan(&u.ID); err != nil {
		t.Fatalf("Error setting up (insert user): %s", err)
	}
	if u.Email != nil {
		query = `
		INSERT INTO emails (userID, email, validated, createDate)
		 VALUES ($1, $2, $3, CURRENT_TIMESTAMP());
		 `
		if _, err := db.Exec(query, u.ID, u.Email.Value,
			u.Email.Verified); err != nil {
			t.Fatalf("Error seting up (inserting email): %s", err)
		}
	}
	if u.UserName != "" {
		query = `
		INSERT INTO userNames (userID, userName, createDate)
		 VALUES ($1, $2, CURRENT_TIMESTAMP());
		 `
		if _, err := db.Exec(query, u.ID, u.UserName); err != nil {
			t.Fatalf("Error seting up (inserting username): %s", err)
		}
	}
	if u.Phone != nil {
		query = `
		INSERT INTO phones (userID, phone, validated, createDate)
		 VALUES ($1, $2, $3, CURRENT_TIMESTAMP());
		 `
		if _, err := db.Exec(query, u.ID, u.Phone.Value,
			u.Phone.Verified); err != nil {
			t.Fatalf("Error seting up (inserting phone) %s", err)
		}
	}
	if u.OAuth != nil {
		query = `
		INSERT INTO appUserIDs (userID, appUserID, appName, createDate)
		 VALUES ($1, $2, $3, CURRENT_TIMESTAMP());
		 `
		if _, err := db.Exec(query, u.ID, u.OAuth.AppUserID,
			u.OAuth.AppName); err != nil {
			t.Fatalf("Error seting up (inserting appUserID): %s", err)
		}
	}
}

func assertUsersEqual(act *authms.User, exp *authms.User, t *testing.T) {
	if reflect.DeepEqual(act, exp) {
		return
	}
	if exp == nil {
		if act != nil {
			t.Errorf("Expected nil but got %+v\n", act)
		}
		return
	} else {
		if act == nil {
			t.Errorf("Expected a value %+v but got nil\n", exp)
		}
		return
	}
	if !reflect.DeepEqual(act.OAuth, exp.OAuth) {
		t.Errorf("Expected oauth %+v but got %+v", act.OAuth, exp.OAuth)
	}
	if !reflect.DeepEqual(act.Phone, exp.Phone) {
		t.Errorf("Expected phone %+v but got %+v", act.Phone, exp.Phone)
	}
	if !reflect.DeepEqual(act.Email, exp.Email) {
		t.Errorf("Expected email %+v but got %+v", act.Email, exp.Email)
	}
	if act.ID != exp.ID {
		t.Errorf("Expected id %d but got %d", exp.ID, act.ID)
	}
	if act.UserName != exp.UserName {
		t.Errorf("Expected UserName %d but got %d", exp.UserName, act.UserName)
	}
}