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
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/go-commons/errors"
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
	if err := buildGcloud(); err != nil {
		log.Fatalf("build GCloud error: %v", err)
	}
}

func installVars() error {
	content := `#!/usr/bin/env bash
NAME="` + config.Name + `"
VERSION="` + config.Version + `"
DESCRIPTION="` + config.Description + `"
CANONICAL_NAME="` + config.CanonicalName + `"
CONF_DIR="` + config.DefaultConfDir() + `"
CONF_FILE="` + config.DefaultConfPath() + `"
INSTALL_DIR="` + config.DefaultInstallDir() + `"
INSTALL_FILE="` + config.DefaultInstallPath() + `"
UNIT_NAME="` + config.DefaultSysDUnitName + `"
UNIT_FILE="` + config.DefaultSysDUnitFilePath() + `"
TPL_DIR="` + config.DefaultTplDir() + `"
EMAIL_INVITE_TPL="` + config.DefaultEmailInviteTpl() + `"
PHONE_INVITE_TPL="` + config.DefaultPhoneInviteTpl() + `"
EMAIL_RESET_PASS_TPL="` + config.DefaultEmailResetPassTpl() + `"
PHONE_RESET_PASS_TPL="` + config.DefaultPhoneResetPassTpl() + `"
EMAIL_VERIFY_TPL="` + config.DefaultEmailVerifyTpl() + `"
PHONE_VERIFY_TPL="` + config.DefaultPhoneVerifyTpl() + `"
`
	return ioutil.WriteFile("install/vars.sh", []byte(content), 0755)
}

func build(goos, goarch, goarm string) error {
	args := []string{"build", "-o", "bin/app", "./cmd/microservice"}
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

func buildGcloud() error {
	confDir := config.DefaultConfDir("cmd", "gcloud", "conf")

	if err := os.MkdirAll(confDir, 0755); err != nil {
		return errors.Newf("create conf dir: %v", err)
	}

	if err := os.MkdirAll(config.DefaultTplDir(), 0755); err != nil {
		return errors.Newf("create templates dir: %v", err)
	}

	err := copyIfDestNotExists(path.Join("install", "conf.yml"), config.DefaultConfPath())
	if err != nil {
		return err
	}
	if err := cleanGCloudConfFile(); err != nil {
		return errors.Newf("clean gcloud config file: %v", err)
	}

	err = copyIfDestNotExists(path.Join("install", "invitation_email.html"), config.DefaultEmailInviteTpl())
	if err != nil {
		return errors.Newf("copy email invite template: %v", err)
	}

	err = copyIfDestNotExists(path.Join("install", "invitation_sms.tpl"), config.DefaultPhoneInviteTpl())
	if err != nil {
		return errors.Newf("copy SMS invite template: %v", err)
	}

	err = copyIfDestNotExists(path.Join("install", "reset_pass_email.html"), config.DefaultEmailResetPassTpl())
	if err != nil {
		return errors.Newf("copy email reset password template: %v", err)
	}

	err = copyIfDestNotExists(path.Join("install", "reset_pass_sms.tpl"), config.DefaultPhoneResetPassTpl())
	if err != nil {
		return errors.Newf("copy phone reset password template: %v", err)
	}

	err = copyIfDestNotExists(path.Join("install", "verify_email.html"), config.DefaultEmailVerifyTpl())
	if err != nil {
		return errors.Newf("copy email verification template: %v", err)
	}

	err = copyIfDestNotExists(path.Join("install", "verify_sms.tpl"), config.DefaultPhoneVerifyTpl())
	if err != nil {
		return errors.Newf("copy phone verification template: %v", err)
	}

	return nil
}

func cleanGCloudConfFile() error {
	newPath := config.DefaultConfPath()
	confContent, err := ioutil.ReadFile(config.DefaultConfPath())
	if err != nil {
		return errors.Newf("read file for transform: %v", err)
	}
	confContentClean := bytes.Replace(confContent, []byte("/etc/authms/"), []byte(""), -1)
	err = ioutil.WriteFile(newPath, confContentClean, 0644)
	if err != nil {
		return errors.Newf("write transformed file: %v", err)
	}
	return nil
}

// copyIfDestNotExists copyis the from file into dest file if dest does not exists.
// see copyFile for Notes.
func copyIfDestNotExists(from, dest string) error {
	_, err := os.Stat(dest)
	if err == nil {
		fmt.Printf("'%s' ignored, already exists\n", dest)
		return nil
	}
	if !os.IsNotExist(err) {
		return errors.Newf("stat: %v", err)
	}
	return copyFile(from, dest)
}

// copyFile copies the from file into dest file
// NOTE: Will consume too much memory if from is a large file!
func copyFile(from, dest string) error {
	data, err := ioutil.ReadFile(from)
	if err != nil {
		return errors.Newf("read from-file: %v", err)
	}
	err = ioutil.WriteFile(dest, data, 0644)
	if err != nil {
		return errors.Newf("write dest-file: %v", err)
	}
	return nil
}
