package verification

import (
	"fmt"
	"regexp"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/errors"
)

const timeFormat = time.RFC3339

type SMSer interface {
	IsNotImplementedError(error) bool
	SMS(toPhone, message string) error
}

type Claims struct {
	jwt.StandardClaims
	Code  string
	Phone string
}

type Tokener interface {
	Generate(claims jwt.Claims) (string, error)
	Validate(token string, claims jwt.Claims) (*jwt.Token, error)
}

type SecureRandomer interface {
	SecureRandomString(length int) ([]byte, error)
}

type Config interface {
	MessageFormat() string
	ValidityPeriod() time.Duration
}

type Verifier struct {
	smser     SMSer
	tokener   Tokener
	generator SecureRandomer
	config    Config
	errors.NotImplErrCheck
}

func New(c Config, s SMSer, sr SecureRandomer, t Tokener) (*Verifier, error) {
	if s == nil {
		return nil, errors.New("SMSer was nil")
	}
	if sr == nil {
		return nil, errors.New("SecureRandomer was nil")
	}
	if t == nil {
		return nil, errors.New("Tokener was nil")
	}
	if c == nil {
		return nil, errors.New("Config was nil")
	}
	if err := testConfig(c); err != nil {
		return nil, err
	}
	return &Verifier{smser: s, tokener: t, generator: sr, config: c}, nil
}

func (v *Verifier) SendSMSCode(toPhone string) (*authms.SMSVerificationStatus, error) {
	codeB, err := v.generator.SecureRandomString(4)
	if err != nil {
		return nil, errors.Newf("error generating SMS code: %v", err)
	}
	issue := time.Now()
	expiry := time.Now().Add(v.config.ValidityPeriod())
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issue.Unix(),
			ExpiresAt: expiry.Unix(),
		},
		Code:  string(codeB),
		Phone: toPhone,
	}
	token, err := v.tokener.Generate(claims)
	if err != nil {
		return nil, errors.Newf("error generting SMS token: %v", err)
	}
	smsBody := fmt.Sprintf(v.config.MessageFormat(), codeB)
	if err = v.smser.SMS(toPhone, smsBody); err != nil {
		if v.smser.IsNotImplementedError(err) {
			return nil, errors.NewNotImplementedf("%v", err)
		}
		return nil, err
	}
	return &authms.SMSVerificationStatus{
		Token:     token,
		Phone:     toPhone,
		ExpiresAt: expiry.Format(timeFormat),
		Verified:  false,
	}, nil
}

func (v *Verifier) VerifySMSCode(r *authms.SMSVerificationCodeRequest) (*authms.SMSVerificationStatus, error) {
	if r == nil {
		return nil, errors.New("request was empty")
	}
	clms := Claims{}
	_, err := v.tokener.Validate(r.SmsToken, &clms)
	if err != nil {
		return nil, errors.NewClientf("invalid token: %v", err)
	}
	if clms.Code != r.Code {
		return nil, errors.NewClient("invalid SMS code")
	}
	return &authms.SMSVerificationStatus{
		Phone:    clms.Phone,
		Verified: true,
	}, nil
}

func testConfig(c Config) error {
	r, err := regexp.Compile("%s")
	if err != nil {
		return errors.Newf("error compiling message tester regex: %v", err)
	}
	formatters := r.FindAllString(c.MessageFormat(), -1)
	if len(formatters) != 1 {
		return errors.Newf("Expected 1 '%%s' formatter but got %d", len(formatters))
	}
	return nil
}
