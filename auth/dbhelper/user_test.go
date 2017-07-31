package dbhelper_test

import (
	"database/sql"
	"github.com/tomogoma/authms/auth/dbhelper"
	"github.com/tomogoma/authms/proto/authms"
	"reflect"
	"testing"
)

type UpdateTestCase struct {
	Desc string
	User *authms.User
}

var completeUserAppName = "test-app"

func TestModel_SaveUser(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	m := newModel(t)
	usr := &authms.User{
		OAuths: map[string]*authms.OAuth{
			completeUserAppName: {
				AppName:   completeUserAppName,
				AppUserID: "test-user-id",
				AppToken:  "test-app.test-user-id.test-token",
				Verified:  true,
			},
		},
	}
	if err := m.SaveUser(usr); err != nil {
		t.Fatalf("model.Save(): %s", err)
	}
	dbUsr := fetchUser(usr.ID, t)
	assertUsersEqual(dbUsr, usr, t)
}

func TestModel_SaveUser_duplicate(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	m := newModel(t)
	usr := &authms.User{
		OAuths: map[string]*authms.OAuth{
			completeUserAppName: {
				AppName:   completeUserAppName,
				AppUserID: "test-user-id",
				AppToken:  "test-app.test-user-id.test-token",
				Verified:  true,
			},
		},
	}
	err := m.SaveUser(usr)
	if err != nil {
		t.Fatalf("model.Save(): %s", err)
	}
	if m.IsDuplicateError(err) {
		t.Error("Expected no duplicate error assigned on nil error")
	}
	err = m.SaveUser(usr)
	if !m.IsDuplicateError(err) {
		t.Errorf("Expected the error %v to be a duplicate error", err)
	}
}

func TestModel_UpdateAppUserID(t *testing.T) {
	bareBoneUser := &authms.User{
		UserName: "test-username",
		Password: "test-password",
	}
	cmpltUsr := completeUser()
	appName := completeUserAppName
	appToken := cmpltUsr.OAuths[appName].AppToken
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
			if expUsr.OAuths == nil {
				expUsr.OAuths = map[string]*authms.OAuth{
					appName: {
						AppName:  appName,
						AppToken: appToken,
						Verified: true,
					},
				}
			}
			expUsr.OAuths[appName].AppUserID = "test-user-id-updated"
			err := m.UpdateAppUserID(tc.User.ID, expUsr.OAuths[appName])
			if err != nil {
				t.Errorf("%s - model.UpdateAppUserID(): %s", tc.Desc, err)
				return
			}
			dbUsr := fetchUser(expUsr.ID, t)
			assertUsersEqual(dbUsr, expUsr, t)
		}()
	}
}

func TestDBHelper_UserExists(t *testing.T) {
	type TestCase struct {
		Desc       string
		InsertUser *authms.User
		TestUser   *authms.User
		ExpExists  bool
	}
	cmpltUsr := completeUser()
	tcs := []TestCase{
		{
			Desc:       "Not exist",
			InsertUser: nil,
			TestUser:   cmpltUsr,
			ExpExists:  false,
		},
		{
			Desc:       "UserName Exists",
			InsertUser: cmpltUsr,
			TestUser:   &authms.User{UserName: cmpltUsr.UserName},
			ExpExists:  true,
		},
		{
			Desc:       "Phone Exists",
			InsertUser: cmpltUsr,
			TestUser:   &authms.User{Phone: cmpltUsr.Phone},
			ExpExists:  true,
		},
		{
			Desc:       "Email Exists",
			InsertUser: cmpltUsr,
			TestUser:   &authms.User{Email: cmpltUsr.Email},
			ExpExists:  true,
		},
		{
			Desc:       "OAuth Exists",
			InsertUser: cmpltUsr,
			TestUser:   &authms.User{OAuths: cmpltUsr.OAuths},
			ExpExists:  true,
		},
	}
	for _, tc := range tcs {
		func() {
			setUp(t)
			defer tearDown(t)
			m := newModel(t)
			if tc.InsertUser != nil {
				insertUser(tc.InsertUser, t)
			}
			usrID, err := m.UserExists(tc.TestUser)
			if err != nil {
				t.Errorf("%s - dbhelper.UserExists(): %v",
					tc.Desc, err)
				return
			}
			if tc.ExpExists && usrID != tc.InsertUser.ID {
				t.Errorf("%s - expected existing userID %d "+
					"but got %d", tc.Desc, tc.InsertUser.ID,
					usrID)
			}
			if !tc.ExpExists && usrID != -1 {
				t.Errorf("%s - expected non-existing userID -1"+
					" but got %d", tc.Desc, usrID)
			}
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
			expUsr.Phone.Verified = true
			err := m.UpdatePhone(tc.User.ID, expUsr.Phone)
			if err != nil {
				t.Errorf("%s - model.UpdatePhone(): %s", tc.Desc, err)
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
			expUsr.Email.Verified = true
			err := m.UpdateEmail(tc.User.ID, expUsr.Email)
			if err != nil {
				t.Errorf("%s - model.UpdateEmail(): %s", tc.Desc, err)
				return
			}
			dbUsr := fetchUser(expUsr.ID, t)
			assertUsersEqual(dbUsr, expUsr, t)
		}()
	}
}

func TestModel_UpdateUserName(t *testing.T) {
	bareBoneUser := &authms.User{
		Phone:    &authms.Value{Value: "+254012345678"},
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
			err := m.UpdateUserName(tc.User.ID, expUsr.UserName)
			if err != nil {
				t.Errorf("%s - model.UpdateUserName(): %s", tc.Desc, err)
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
			expErr: dbhelper.ErrorPasswordMismatch},
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
				t.Errorf("%s - model.UpdatePassword() expected"+
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
				t.Fatalf("%s - error retrieving password for"+
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
	actUser, err := m.GetByAppUserID(expUsr.OAuths[completeUserAppName].AppName,
		expUsr.OAuths[completeUserAppName].AppUserID)
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

func completeUser() *authms.User {
	return &authms.User{
		UserName: "test-username",
		Email: &authms.Value{
			Value:    "test@email.com",
			Verified: false,
		},
		Phone: &authms.Value{
			Value:    "+254712345678",
			Verified: false,
		},
		OAuths: map[string]*authms.OAuth{
			completeUserAppName: {
				AppName:   completeUserAppName,
				AppUserID: "test-user-id",
				AppToken:  "test-app.test-user-id.test-token",
				Verified:  false,
			},
		},
		Password: "test-password",
	}
}

func fetchUser(userID int64, t *testing.T) *authms.User {
	db := getDB(t)
	query := `
	SELECT users.id, userNames.userName, phones.phone, phones.validated,
		  emails.email, emails.validated, appUserIDs.appName,
		  appUserIDs.appUserID, appUserIDs.validated
		FROM users
		LEFT JOIN userNames ON users.id=userNames.userID
		LEFT JOIN phones ON users.id=phones.userID
		LEFT JOIN emails ON users.id=emails.userID
		LEFT JOIN appUserIDs ON users.id=appUserIDs.userID
		WHERE users.id=$1`
	dbUsr := &authms.User{
		Email:  &authms.Value{},
		Phone:  &authms.Value{},
		OAuths: make(map[string]*authms.OAuth),
	}
	var dbUserName, dbPhone, dbEmail, dbAppName, dbAppUsrID sql.NullString
	var dbPhoneValidated, dbEmailValidated, dbAppValidated sql.NullBool
	err := db.QueryRow(query, userID).Scan(&dbUsr.ID, &dbUserName, &dbPhone,
		&dbPhoneValidated, &dbEmail, &dbEmailValidated, &dbAppName,
		&dbAppUsrID, &dbAppValidated)
	dbUsr.UserName = dbUserName.String
	dbUsr.Phone.Value = dbPhone.String
	dbUsr.Phone.Verified = dbPhoneValidated.Bool
	dbUsr.Email.Value = dbEmail.String
	dbUsr.Email.Verified = dbEmailValidated.Bool
	oa := &authms.OAuth{AppName: dbAppName.String, AppUserID: dbAppUsrID.String,
		Verified: dbAppValidated.Bool}
	dbUsr.OAuths[oa.AppName] = oa
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
	for _, oauth := range u.OAuths {
		query = `
		INSERT INTO appUserIDs (userID, appUserID, appName, createDate)
		 VALUES ($1, $2, $3, CURRENT_TIMESTAMP());
		 `
		if _, err := db.Exec(query, u.ID, oauth.AppUserID,
			oauth.AppName); err != nil {
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
	} else if act == nil {
		t.Errorf("Expected a value %+v but got nil\n", exp)
		return
	}
	for appName, expOAuth := range exp.OAuths {
		actOAuth := act.OAuths[appName]
		if !oAuthEqual(actOAuth, expOAuth) {
			t.Errorf("Expected oauth %+v but got %+v", expOAuth, actOAuth)
		}
	}
	if !valuesEqual(act.Phone, exp.Phone) {
		t.Errorf("Expected phone %+v but got %+v", exp.Phone, act.Phone)
	}
	if !valuesEqual(act.Email, exp.Email) {
		t.Errorf("Expected email %+v but got %+v", exp.Email, act.Email)
	}
	if act.ID != exp.ID {
		t.Errorf("Expected id %d but got %d", exp.ID, act.ID)
	}
	if act.UserName != exp.UserName {
		t.Errorf("Expected UserName %d but got %d", exp.UserName, act.UserName)
	}
}

func valuesEqual(act, exp *authms.Value) bool {
	if exp == nil {
		return act == nil || act.Value == ""
	} else if act == nil {
		return false
	}
	return act.Value == exp.Value && act.Verified == exp.Verified
}

func oAuthEqual(act, exp *authms.OAuth) bool {
	if exp == nil {
		return act == nil ||
			(act.AppName == "" && act.AppUserID == "" && act.Verified == false)
	} else if act == nil {
		return false
	}
	return act.AppName == exp.AppName && act.AppUserID == exp.AppUserID &&
		act.Verified == exp.Verified
}
