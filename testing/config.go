package testing

import (
	"flag"
	"testing"

	"github.com/tomogoma/authms/config"
	config2 "github.com/tomogoma/go-commons/config"
)

var confPath = flag.String("conf", config.DefaultConfPath, "/path/to/config")

func init() {
	flag.Parse()
}

func ReadConfig(t *testing.T) config.General {
	conf := config.General{}
	if err := config2.ReadYamlConfig(*confPath, &conf); err != nil {
		t.Fatalf("Error setting up: read config: %v", err)
	}
	return conf
}
