package user

import (
	"errors"

	"fmt"

	"github.com/tomogoma/authms/auth/model/history"
)

var ErrorInvalidID = errors.New("the user ID provided was invalid")
var ErrorEmptyToken = errors.New("the token provided was empty")
var ErrorInvalidOAuth = errors.New("the oauth provided was invalid")
var ErrorEmptyEmail = errors.New("email was empty")
var ErrorEmptyPhone = errors.New("phone was empty")
var ErrorEmptyUserName = errors.New("username was empty")
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

func NewByUserName(uName, pass string, hashF HashFunc) (*user, error) {
	if uName == "" {
		return nil, ErrorEmptyUserName
	}
	passHB, err := hashPassword(pass, hashF)
	if err != nil {
		return nil, err
	}
	return &user{
		userName: uName,
		phone:    new(value),
		email:    new(value),
		app:      new(app),
		password: passHB,
	}, nil
}

func NewByPhone(phoneNo, pass string, hashF HashFunc) (*user, error) {
	if phoneNo == "" {
		return nil, ErrorEmptyPhone
	}
	passHB, err := hashPassword(pass, hashF)
	if err != nil {
		return nil, err
	}
	return &user{
		phone:    &value{value: phoneNo},
		email:    new(value),
		app:      new(app),
		password: passHB,
	}, nil
}

func NewByEmail(email, pass string, hashF HashFunc) (*user, error) {
	if email == "" {
		return nil, ErrorEmptyEmail
	}
	passHB, err := hashPassword(pass, hashF)
	if err != nil {
		return nil, err
	}
	return &user{
		phone:    new(value),
		app:      new(app),
		email:    &value{value: email},
		password: passHB,
	}, nil
}

func NewByOAuth(oauth App, gen PasswordGenerator, hashF HashFunc) (*user, error) {
	if !appIsFilled(oauth) {
		return nil, ErrorInvalidOAuth
	}
	if gen == nil {
		return nil, ErrorNilPasswordGenerator
	}
	passB, err := gen.SecureRandomString(36)
	if err != nil {
		return nil, fmt.Errorf("failed generate secure random string: %s", err)
	}
	pass := string(passB)
	if pass == "" {
		return nil, errors.New("failed generate secure random string: got empty")
	}
	passHB, err := hashPassword(pass, hashF)
	if err != nil {
		return nil, err
	}
	a := &app{
		name:      oauth.Name(),
		userID:    oauth.UserID(),
		validated: oauth.Validated(),
	}
	return &user{
		phone:    new(value),
		email:    new(value),
		app:      a,
		password: passHB,
	}, nil
}

func NewByToken(id int, token string) (*user, error) {
	if id < 1 {
		return nil, ErrorInvalidID
	}
	if token == "" {
		return nil, ErrorEmptyToken
	}
	return &user{
		id:    id,
		phone: new(value),
		app:   new(app),
		email: new(value),
		token: token,
	}, nil
}

func hashPassword(pass string, hashF HashFunc) ([]byte, error) {
	if pass == "" {
		return nil, ErrorEmptyPassword
	}
	if hashF == nil {
		return nil, ErrorNilHashFunc
	}
	return hashF(pass)
}
