package config

import "time"

// Compile time constants that should not be configurable
// during runtime.
const (
	Name                = "authms"
	Version             = "v1"
	Description         = "Authentication Micro-Service"
	CanonicalName       = Name + Version
	RPCNamePrefix       = ""
	CanonicalRPCName    = RPCNamePrefix + CanonicalName
	WebNamePrefix       = "go.micro.web."
	CanonicalWebName    = WebNamePrefix + CanonicalName
	DefaultSysDUnitName = CanonicalName + ".service"

	DefaultInstallDir        = "/usr/local/bin"
	DefaultInstallPath       = DefaultInstallDir + "/" + CanonicalName
	DefaultSysDUnitFilePath  = "/etc/systemd/system/" + DefaultSysDUnitName
	DefaultConfDir           = "/etc/" + Name
	DefaultConfPath          = DefaultConfDir + "/" + CanonicalName + ".conf.yml"
	DefaultTplDir            = DefaultConfDir + "/templates"
	DefaultEmailInviteTpl    = DefaultTplDir + "/" + CanonicalName + "_email_invite.html"
	DefaultPhoneInviteTpl    = DefaultTplDir + "/" + CanonicalName + "_phone_invite.tpl"
	DefaultEmailResetPassTpl = DefaultTplDir + "/" + CanonicalName + "_email_reset_pass.html"
	DefaultPhoneResetPassTpl = DefaultTplDir + "/" + CanonicalName + "_phone_reset_pass.tpl"
	DefaultEmailVerifyTpl    = DefaultTplDir + "/" + CanonicalName + "_email_verify.html"
	DefaultPhoneVerifyTpl    = DefaultTplDir + "/" + CanonicalName + "_phone_verify.tpl"

	SMSAPITwilio         = "twilio"
	SMSAPIAfricasTalking = "africasTalking"
	SMSAPIMessageBird    = "messageBird"

	TimeFormat = time.RFC3339

	APIKeyLength = 56
)
