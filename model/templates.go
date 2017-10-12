package model

type InvitationTemplate struct {
	URLToken string
	AppName  string
}

type VerificationTemplate struct {
	URLToken string
	Code     string
	AppName  string
}
