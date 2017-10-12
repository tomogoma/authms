package config

import (
	"path"
	"time"
)

// Compile time constants that should not be configurable
// during runtime.
const (
	Name                = "authms"
	Version             = "v1"
	Description         = "Authentication Micro-Service"
	CanonicalName       = Name + Version
	RPCNamePrefix       = ""
	CanonicalRPCName    = RPCNamePrefix + CanonicalName
	WebNamePrefix       = "go.micro.api." + Version + "."
	WebRootURL          = "/" + Version + "/" + Name
	CanonicalWebName    = WebNamePrefix + Name
	DefaultSysDUnitName = CanonicalName + ".service"

	SMSAPITwilio         = "twilio"
	SMSAPIAfricasTalking = "africasTalking"
	SMSAPIMessageBird    = "messageBird"

	TimeFormat = time.RFC3339

	APIKeyLength = 56
)

var (
	// FIXME Probably won't work for none-unix systems!
	defaultInstallDir       = path.Join("/usr", "local", "bin")
	defaultSysDUnitFilePath = path.Join("/etc", "systemd", "system", DefaultSysDUnitName)
	sysDConfDir             = path.Join("/etc", Name)
	defaultConfDir          = sysDConfDir
)

func DefaultInstallDir() string {
	return defaultInstallDir
}

func DefaultInstallPath() string {
	return path.Join(defaultInstallDir, CanonicalName)
}

func DefaultSysDUnitFilePath() string {
	return defaultSysDUnitFilePath
}

// DefaultConfDir sets the value of the conf dir to use and returns it.
// It falls back to default - sysDConfDir() - if newPSegments has zero len.
func DefaultConfDir(newPSegments ...string) string {
	if len(newPSegments) == 0 {
		defaultConfDir = sysDConfDir
	} else {
		defaultConfDir = path.Join(newPSegments...)
	}
	return defaultConfDir
}

func DefaultConfPath() string {
	return path.Join(defaultConfDir, CanonicalName+".conf.yml")
}

func DefaultTplDir() string {
	return path.Join(defaultConfDir, "templates")
}

func DefaultEmailInviteTpl() string {
	return path.Join(DefaultTplDir(), CanonicalName+"_email_invite.html")
}

func DefaultPhoneInviteTpl() string {
	return path.Join(DefaultTplDir(), CanonicalName+"_phone_invite.tpl")
}

func DefaultEmailResetPassTpl() string {
	return path.Join(DefaultTplDir(), CanonicalName+"_email_reset_pass.html")
}

func DefaultPhoneResetPassTpl() string {
	return path.Join(DefaultTplDir(), CanonicalName+"_phone_reset_pass.tpl")
}

func DefaultEmailVerifyTpl() string {
	return path.Join(DefaultTplDir(), CanonicalName+"_email_verify.html")
}

func DefaultPhoneVerifyTpl() string {
	return path.Join(DefaultTplDir(), CanonicalName+"_phone_verify.tpl")
}
