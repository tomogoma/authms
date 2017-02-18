package user_test

import (
	"testing"
	"github.com/tomogoma/authms/auth/model/user"
	"io/ioutil"
	"flag"
	"gopkg.in/yaml.v2"
	"github.com/tomogoma/authms/auth/password"
	"github.com/tomogoma/authms/auth/hash"
	"github.com/tomogoma/go-commons/errors"
	"github.com/tomogoma/authms/auth/model/helper"
	token_config "github.com/tomogoma/authms/auth/model/token"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/authms/proto/authms"
	"database/sql"
	"reflect"
	"github.com/tomogoma/go-commons/auth/token"
)

type Config struct {
	Database helper.DSN    `json:"database,omitempty"`
	Token    token_config.Config  `json:"token,omitempty"`
}

var confFile = flag.String("conf", "/etc/authms/authms.conf.yml", "/path/to/conf.file.yml")
var conf = &Config{}
var hasher = hash.Hasher{}
var tg *password.Generator

func init() {
	flag.Parse()
}

func TestNewModel(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	newModel(t)
}

func TestModel_Save(t *testing.T) {
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
	if err := m.Save(usr); err != nil {
		t.Fatalf("model.Save(): %s", err)
	}
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
	var dbUserName, dbPhone, dbEmail sql.NullString
	var dbPhoneValidated, dbEmailValidated sql.NullBool
	err := db.QueryRow(query, usr.ID).Scan(&dbUsr.ID,
		&dbUserName, &dbPhone, &dbPhoneValidated, &dbEmail,
		&dbEmailValidated, &dbUsr.OAuth.AppName,
		&dbUsr.OAuth.AppUserID)
	dbUsr.UserName = dbUserName.String
	dbUsr.Phone.Value = dbPhone.String
	dbUsr.Phone.Verified = dbPhoneValidated.Bool
	dbUsr.Email.Value = dbEmail.String
	dbUsr.Email.Verified = dbEmailValidated.Bool
	if err != nil {
		t.Fatalf("Error checking db contents for validation: %s", err)
	}
	assertUsersEqual(dbUsr, usr, t)
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
	query = `
	INSERT INTO emails (userID, email, validated, createDate)
	 VALUES ($1, $2, $3, CURRENT_TIMESTAMP());
		 `
	if _, err := db.Exec(query, u.ID, u.Email.Value,
		u.Email.Verified); err != nil {
		t.Fatalf("Error seting up (inserting email): %s", err)
	}
	query = `
	INSERT INTO userNames (userID, userName, createDate)
	 VALUES ($1, $2, CURRENT_TIMESTAMP());
		 `
	if _, err := db.Exec(query, u.ID, u.UserName); err != nil {
		t.Fatalf("Error seting up (inserting username): %s", err)
	}
	query = `
	INSERT INTO phones (userID, phone, validated, createDate)
	 VALUES ($1, $2, $3, CURRENT_TIMESTAMP());
		 `
	if _, err := db.Exec(query, u.ID, u.Phone.Value,
		u.Phone.Verified); err != nil {
		t.Fatalf("Error seting up (inserting phone) %s", err)
	}
	query = `
	INSERT INTO appUserIDs (userID, appUserID, appName, createDate)
	 VALUES ($1, $2, $3, CURRENT_TIMESTAMP());
		 `
	if _, err := db.Exec(query, u.ID, u.OAuth.AppUserID,
		u.OAuth.AppName); err != nil {
		t.Fatalf("Error seting up (inserting appUserID): %s", err)
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

func newModel(t *testing.T) (*user.Model) {
	pg, err := password.NewGenerator(password.AllChars)
	if err != nil {
		t.Fatalf("password.NewGenerator(): %s", err)
	}
	tg, err := token.NewGenerator(conf.Token)
	if err != nil {
		t.Fatalf("token.NewGenerator(): %s", err)
	}
	m, err := user.NewModel(conf.Database, pg, hasher, tg)
	if err != nil {
		t.Fatalf("user.NewModel(): %s", err)
	}
	return m
}

func getDB(t *testing.T) *sql.DB {
	db, err := cockroach.DBConn(conf.Database)
	if err != nil {
		t.Fatalf("unable to tear down: cockroach.DBConn(): %s", err)
	}
	return db
}

func setUp(t *testing.T) {
	if err := readConfig(*confFile, conf); err != nil {
		t.Fatal(err)
	}
	conf.Database.DB = conf.Database.DB + "_test"
}

func tearDown(t *testing.T) {
	db := getDB(t)
	_, err := db.Exec("DROP DATABASE IF EXISTS " + conf.Database.DBName())
	if err != nil {
		t.Fatalf("unable to tear down db: %s", err)
	}
}

func readConfig(confFile string, conf interface{}) error {
	configB, err := ioutil.ReadFile(confFile)
	if err != nil {
		return errors.Newf("unable to read conf file at '%s': %s",
			confFile, err)
	}
	if err := yaml.Unmarshal(configB, conf); err != nil {
		return errors.Newf("unable to unmarshal conf file values for" +
			" file at '%s': %s", confFile, err)
	}
	return nil
}