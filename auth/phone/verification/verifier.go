package verification

import (
	"github.com/tomogoma/go-commons/errors"
	"github.com/tomogoma/authms/proto/authms"
	jwt "github.com/dgrijalva/jwt-go"
	"regexp"
	"fmt"
	"time"
)

const timeFormat = time.RFC3339

type SMSer interface {
	SMS(toPhone, message string) error
}

type Claims  struct {
	jwt.StandardClaims
	Code  string
	Phone string
}

type Tokener interface {
	GenerateWithClaims(claims jwt.Claims) (string, error)
	ValidateClaims(token string, claims jwt.Claims) (*jwt.Token, error)
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
			IssuedAt: issue.Unix(),
			ExpiresAt: expiry.Unix(),
		},
		Code: string(codeB),
		Phone: toPhone,
	}
	token, err := v.tokener.GenerateWithClaims(claims)
	if err != nil {
		return nil, errors.Newf("error generting SMS token: %v", err)
	}
	smsBody := fmt.Sprintf(v.config.MessageFormat(), codeB)
	fmt.Println(smsBody)
	if err = v.smser.SMS(toPhone, smsBody); err != nil {
		return nil, errors.Newf("error sending SMS: %s", err)
	}
	return &authms.SMSVerificationStatus{
		Token: token,
		Phone: toPhone,
		ExpiresAt: expiry.Format(timeFormat),
		Verified: false,
	}, nil
}

func (v *Verifier) VerifySMSCode(r *authms.SMSVerificationCodeRequest) (*authms.SMSVerificationStatus, error) {
	if r == nil {
		return nil, errors.New("request was empty")
	}
	token, err := v.tokener.ValidateClaims(r.Token, &Claims{})
	if err != nil {
		return nil, errors.NewClientf("invalid token: %v", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.NewClient("invalid token claims")
	}
	if claims.Code != r.Code {
		return nil, errors.NewClient("invalid SMS code")
	}
	return &authms.SMSVerificationStatus{
		Phone: claims.Phone,
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