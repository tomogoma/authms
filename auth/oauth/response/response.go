package response

type OAuth interface {
	IsValid() bool
	UserID() string
	AppID() string
	AppName() string
	Scopes() []string
	Metadata() map[string]string
}
