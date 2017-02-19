package model_test

import (
	"testing"
	"github.com/tomogoma/authms/auth/model"
	"io/ioutil"
	"flag"
	"gopkg.in/yaml.v2"
	"github.com/tomogoma/authms/auth/password"
	"github.com/tomogoma/authms/auth/hash"
	"github.com/tomogoma/go-commons/errors"
	"github.com/tomogoma/go-commons/database/cockroach"
	"database/sql"
	"github.com/tomogoma/go-commons/auth/token"
)

type Token struct {
	TknKeyFile string `yaml:"tokenkeyfile"`
}

func (c Token) TokenKeyFile() string {
	return c.TknKeyFile
}

type Config struct {
	Database cockroach.DSN    `json:"database,omitempty"`
	Token    Token  `json:"token,omitempty"`
}

var confFile = flag.String("conf", "/etc/authms/authms.conf.yml", "/path/to/conf.file.yml")
var conf = &Config{}
var hasher = hash.Hasher{}
var tg *token.Generator

func init() {
	flag.Parse()
}

func TestNewModel(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	newModel(t)
}

func newModel(t *testing.T) (*model.Model) {
	pg, err := password.NewGenerator(password.AllChars)
	if err != nil {
		t.Fatalf("password.NewGenerator(): %s", err)
	}
	tg, err = token.NewGenerator(conf.Token)
	if err != nil {
		t.Fatalf("token.NewGenerator(): %s", err)
	}
	m, err := model.New(conf.Database, pg, hasher, tg)
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