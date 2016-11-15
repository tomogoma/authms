package user

import (
	"errors"

	"fmt"

	"github.com/tomogoma/authms/auth/model/history"
)

var ErrorEmptyIdentifier = errors.New("must have at least one login identifier (username|email|appID...)")
var ErrorEmptyPassword = errors.New("password cannot be empty")
var ErrorNilHashFunc = errors.New("hash function cannot be nil")
var ErrorNilPasswordGenerator = errors.New("password generator was nil")

type Validated interface {
	Validated() bool
}

type Valuer interface {
	Validated
	Value() string
}

type PasswordGenerator interface {
	SecureRandomString(length int) ([]byte, error)
}

func valIsFilled(v Valuer) bool {
	return v != nil && v.Value() != ""
}

type User interface {
	ID() int
	UserName() string
	Phone() Valuer
	Email() Valuer
	App() App
	PreviousLogins() []*history.History
	Token(string) string
}

type user struct {
	id             int
	userName       string
	phone          *value
	email          *value
	app            *app
	password       []byte
	previousLogins []*history.History
	token          string
}

func (u *user) ID() int {
	return u.id
}
func (u *user) UserName() string {
	return u.userName
}
func (u *user) Phone() Valuer {
	return u.phone
}
func (u *user) Email() Valuer {
	return u.email
}
func (u *user) App() App {
	return u.app
}
func (u *user) PreviousLogins() []*history.History {
	return u.previousLogins
}
func (u *user) Token(tkn string) string {
	if tkn == "" {
		return u.token
	}
	u.token = tkn
	return u.token
}

func (u *user) SetPreviousLogins(ls ...*history.History) {
	u.previousLogins = ls
}

func New(uName, phoneNo, email, pass string, appUserID App, gen PasswordGenerator, hashF HashFunc) (*user, error) {
	if appIsFilled(appUserID) {
		if uName == "" && phoneNo == "" && email == "" {
			if gen == nil {
				return nil, ErrorNilPasswordGenerator
			}
			gen, err := gen.SecureRandomString(36)
			if err != nil {
				return nil, fmt.Errorf("failed generate secure random string: %s", err)
			}
			pass = string(gen)
		}
	} else if uName == "" && phoneNo == "" && email == "" {
		return nil, ErrorEmptyIdentifier
	}
	if pass == "" {
		return nil, ErrorEmptyPassword
	}
	if hashF == nil {
		return nil, ErrorNilHashFunc
	}
	passHB, err := hashF(pass)
	if err != nil {
		return nil, err
	}
	var a *app
	if appIsFilled(appUserID) {
		a = &app{
			name:      appUserID.Name(),
			userID:    appUserID.UserID(),
			validated: appUserID.Validated(),
		}
	}
	return &user{
		userName: uName,
		phone:    &value{value: phoneNo},
		email:    &value{value: email},
		app:      a,
		password: passHB,
	}, nil
}
