package model

type FacebookResponse interface {
	UserID() string
	AppID() string
	AppName() string
	Scopes() []string
	Metadata() map[string]string
}
