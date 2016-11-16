package user

type App interface {
	Validated
	UserID() string
	Name() string
	Token() string
}

func appIsFilled(a App) bool {
	return a != nil && a.Name() != "" && a.UserID() != ""
}

type app struct {
	userID    string
	name      string
	validated bool
}

func (a *app) UserID() string {
	if a == nil {
		return ""
	}
	return a.userID
}

func (a *app) Name() string {
	if a == nil {
		return ""
	}
	return a.name
}

func (a *app) Token() string {
	return ""
}

func (a *app) Validated() bool {
	if a == nil {
		return false
	}
	return a.validated
}
