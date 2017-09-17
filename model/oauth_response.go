package model

type OAuthResponse interface {
	UserID() string
	AppID() string
	AppName() string
	Scopes() []string
	Metadata() map[string]string
}
