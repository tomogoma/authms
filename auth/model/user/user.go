package user

import (
	"errors"

	"bitbucket.org/tomogoma/auth-ms/auth/model/history"
	"bitbucket.org/tomogoma/auth-ms/auth/model/token"
	"golang.org/x/crypto/bcrypt"
)

var ErrorEmptyUserName = errors.New("uName cannot be empty")
var ErrorEmptyPassword = errors.New("password cannot be empty")
var ErrorNilHashFunc = errors.New("Hash function cannot be nil")

type User interface {
	ID() int
	UserName() string
	FirstName() string
	MiddleName() string
	LastName() string
	PreviousLogins() []*history.History
	Token() token.Token
}

type user struct {
	id             int
	userName       string
	firstName      string
	middleName     string
	lastName       string
	password       []byte
	previousLogins []*history.History
	token          token.Token
}

func (u *user) ID() int                            { return u.id }
func (u *user) UserName() string                   { return u.userName }
func (u *user) FirstName() string                  { return u.firstName }
func (u *user) MiddleName() string                 { return u.middleName }
func (u *user) LastName() string                   { return u.lastName }
func (u *user) PreviousLogins() []*history.History { return u.previousLogins }
func (u *user) Token() token.Token                 { return u.token }

func (u *user) SetPreviousLogins(ls ...*history.History) { u.previousLogins = ls }
func (u *user) SetToken(t token.Token)                   { u.token = t }

type HashFunc func(pass string) ([]byte, error)

func New(uName, fName, mName, lName, pass string, hashF HashFunc) (*user, error) {

	if uName == "" {
		return nil, ErrorEmptyUserName
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

	return &user{
		userName:   uName,
		firstName:  fName,
		middleName: mName,
		lastName:   lName,
		password:   passHB,
	}, nil
}

func Hash(pass string) ([]byte, error) {

	passB := []byte(pass)
	return bcrypt.GenerateFromPassword(passB, bcrypt.DefaultCost)
}
