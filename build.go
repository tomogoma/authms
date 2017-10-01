// +build dev

// build.go automates proper versioning of authms binaries
// and installer scripts.
// Use it like:   go run build.go
// The result binary will be located in bin/app
// You can customize the build with the -goos, -goarch, and
// -goarm CLI options:   go run build.go -goos=windows
//
// This program is NOT required to build authms from source
// since it is go-gettable. (You can run plain `go build`
// in this directory to get a binary).
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/tomogoma/authms/config"
)

func main() {
	var goos, goarch, goarm string
	var help bool
	flag.StringVar(&goos, "goos", "",
		"GOOS\tThe operating system for which to compile\n"+
			"\t\tExamples are linux, darwin, windows, netbsd.")
	flag.StringVar(&goarch, "goarch", "",
		"GOARCH\tThe architecture, or processor, for which to compile code.\n"+
			"\t\tExamples are amd64, 386, arm, ppc64.")
	flag.StringVar(&goarm, "goarm", "",
		"GOARM\tFor GOARCH=arm, the ARM architecture for which to compile.\n"+
			"\t\tValid values are 5, 6, 7.")
	flag.BoolVar(&help, "help", false, "Show this help message")
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}
	if err := build(goos, goarch, goarm); err != nil {
		log.Fatalf("build error: %v", err)
	}
	if err := installVars(); err != nil {
		log.Fatalf("write installer script error: %v", err)
	}
}

func installVars() error {
	content := `#!/usr/bin/env bash
NAME="` + config.Name + `"
VERSION="` + config.Version + `"
DESCRIPTION="` + config.Description + `"
CANONICAL_NAME="` + config.CanonicalName + `"
CONF_DIR="` + config.DefaultConfDir + `"
CONF_FILE="` + config.DefaultConfPath + `"
INSTALL_DIR="` + config.DefaultInstallDir + `"
INSTALL_FILE="` + config.DefaultInstallPath + `"
UNIT_NAME="` + config.DefaultSysDUnitName + `"
UNIT_FILE="` + config.DefaultSysDUnitFilePath + `"
TPL_DIR="` + config.DefaultTplDir + `"
EMAIL_INVITE_TPL="` + config.DefaultEmailInviteTpl + `"
PHONE_INVITE_TPL="` + config.DefaultPhoneInviteTpl + `"
EMAIL_RESET_PASS_TPL="` + config.DefaultEmailResetPassTpl + `"
PHONE_RESET_PASS_TPL="` + config.DefaultPhoneResetPassTpl + `"
`
	return ioutil.WriteFile("install/vars.sh", []byte(content), 0755)
}

func build(goos, goarch, goarm string) error {
	args := []string{"build", "-o", "bin/app"}
	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = os.Environ()
	for _, env := range []string{
		"GOOS=" + goos,
		"GOARCH=" + goarch,
		"GOARM=" + goarm,
	} {
		cmd.Env = append(cmd.Env, env)
	}
	return cmd.Run()
}
