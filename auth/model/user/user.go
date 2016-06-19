package user

import (
	"errors"

	"bitbucket.org/tomogoma/auth-ms/auth/model/details/login"
	"bitbucket.org/tomogoma/auth-ms/auth/model/token"
	"golang.org/x/crypto/bcrypt"
)

var ErrorEmptyUserName = errors.New("uName cannot be empty")
var ErrorEmptyPassword = errors.New("password cannot be empty")

type User struct {
	id             int
	userName       string
	firstName      string
	middleName     string
	lastName       string
	password       []byte
	previousLogins []login.LoginDetails
	token          *token.Token
}

func (u *User) ID()             { return u.id }
func (u *User) UserName()       { return u.userName }
func (u *User) FirstName()      { return u.firstName }
func (u *User) MiddleName()     { return u.middleName }
func (u *User) LastName()       { return u.lastName }
func (u *User) PreviousLogins() { return u.previousLogins }
func (u *User) Token()          { return u.token }

func (u *User) SetPreviousLogins(ls []login.LoginDetails) { u.previousLogins = ls }
func (u *User) SetToken(t *token.Token)                   { u.token = t }

func New(uName, fName, mName, lName, pass string) (*User, error) {

	if uName == "" {
		return nil, ErrorEmptyUserName
	}

	if pass == "" {
		return nil, ErrorEmptyPassword
	}

	passHB, err := getHash(pass)
	if err != nil {
		return nil, err
	}

	return &User{
		userName:  uName,
		firstName: fName,
		lastName:  lName,
		password:  passHB,
	}, nil
}

func getHash(pass string) ([]byte, error) {

	passB := []byte(pass)
	return bcrypt.GenerateFromPassword(passB, bcrypt.DefaultCost)
}
