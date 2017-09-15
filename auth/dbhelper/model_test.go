package dbhelper_test

import (
	"database/sql"
	"flag"
	"testing"

	"github.com/tomogoma/authms/auth/dbhelper"
	"github.com/tomogoma/authms/auth/hash"
	"github.com/tomogoma/authms/auth/password"
	"github.com/tomogoma/go-commons/config"
	"github.com/tomogoma/go-commons/database/cockroach"
)

type Token struct {
	TknKeyFile string `yaml:"tokenkeyfile"`
}

func (c Token) TokenKeyFile() string {
	return c.TknKeyFile
}

type Config struct {
	Database cockroach.DSN `json:"database,omitempty"`
	Token    Token         `json:"token,omitempty"`
}

var confFile = flag.String("conf", "/etc/authms/authms.conf.yml", "/path/to/conf.file.yml")
var conf = &Config{}
var hasher = hash.Hasher{}

func init() {
	flag.Parse()
}

func TestNewModel(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	newModel(t)
}

func newModel(t *testing.T) *dbhelper.DBHelper {
	pg, err := password.NewGenerator(password.AllChars)
	if err != nil {
		t.Fatalf("password.NewGenerator(): %s", err)
	}
	m, err := dbhelper.New(conf.Database, pg, hasher)
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
	if err := config.ReadYamlConfig(*confFile, conf); err != nil {
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
