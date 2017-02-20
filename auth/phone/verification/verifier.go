package verification

import (
	"github.com/tomogoma/go-commons/errors"
	"github.com/tomogoma/authms/proto/authms"
	jwt "github.com/dgrijalva/jwt-go"
)

type SMSer interface {
	SMS(toPhone, message string) error
}

type Claims  struct {
	authms.SMSVerificationStatus
	jwt.StandardClaims
}

func (c Claims) Valid() error {
	return errors.New("Not yet implemented")
}

type Tokener interface {
	GenerateWithClaims(claims jwt.Claims) (string, error)
	ValidateClaims(token string, claims jwt.Claims) (*jwt.Token, error)
}

type Verifier struct {
	smser   SMSer
	tokener Tokener
}

func New(s SMSer, t Tokener) (*Verifier, error) {
	return nil, errors.New("Not yet implemented")
}

func (v *Verifier) SendSMSCode(toPhone string) (*authms.SMSVerificationStatus, error) {
	return nil, errors.New("Not yet implemented")
}

func (v *Verifier) VerifySMSCode(r *authms.SMSVerificationCodeRequest) (*authms.SMSVerificationStatus, error) {
	return nil, errors.New("Not yet implemented")
}