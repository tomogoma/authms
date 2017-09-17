package verification_test

import (
	"testing"
	"time"

	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/verification"
	"github.com/tomogoma/go-commons/errors"
)

type SMSerMock struct {
	errors.NotImplErrCheck
	SMSCalled bool
	ExpErr    error
}

func (s *SMSerMock) SMS(toPhone, message string) error {
	s.SMSCalled = true
	return s.ExpErr
}

type TokenerMock struct {
	ExpToken                 string
	ExpErr                   error
	ExpJwt                   *jwt.Token
	GenerateWithClaimsCalled bool
	ValidateClaimsCalled     bool
	ExpCode                  string
}

func (t *TokenerMock) Generate(claims jwt.Claims) (string, error) {
	t.GenerateWithClaimsCalled = true
	return t.ExpToken, t.ExpErr
}
func (t *TokenerMock) Validate(token string, claims jwt.Claims) (*jwt.Token, error) {
	t.ValidateClaimsCalled = true
	cl, ok := claims.(*verification.Claims)
	if ok {
		cl.Code = t.ExpCode
	}
	if t.ExpJwt != nil {
		t.ExpJwt.Claims = cl
	}
	return t.ExpJwt, t.ExpErr
}

type SecureRandomerMock struct {
	ExpString                string
	ExpErr                   error
	SecureRandomStringCalled bool
}

func (sr *SecureRandomerMock) SecureRandomBytes(length int) ([]byte, error) {
	sr.SecureRandomStringCalled = true
	return []byte(sr.ExpString), sr.ExpErr
}

type VerifierTestCase struct {
	Desc           string
	ExpErr         bool
	ExpNotImpl     bool
	Phone          string
	CodeReq        *authms.SMSVerificationCodeRequest
	SMSer          *SMSerMock
	Tokener        *TokenerMock
	SecureRandomer *SecureRandomerMock
	SMSFormat      string
	CodeValidity   time.Duration
	ExpVerStatus   *authms.SMSVerificationStatus
}

func validDependencies() VerifierTestCase {
	return VerifierTestCase{
		Desc:   "Valid args",
		ExpErr: false,
		SMSer:  &SMSerMock{ExpErr: nil},
		Tokener: &TokenerMock{
			ExpErr:   nil,
			ExpToken: "some-token",
			ExpJwt:   &jwt.Token{Raw: "some-token"},
		},
		SecureRandomer: &SecureRandomerMock{
			ExpErr: nil, ExpString: "1234",
		},
		CodeValidity: 5 * time.Minute,
		SMSFormat:    "verification code is %s",
	}
}

func TestNew(t *testing.T) {
	tcs := []struct {
		Desc           string
		ExpErr         bool
		ExpNotImpl     bool
		Phone          string
		CodeReq        *authms.SMSVerificationCodeRequest
		SMSer          verification.SMSer
		Tokener        verification.Tokener
		SecureRandomer verification.SecureRandomer
		SMSFormat      string
		CodeValidity   time.Duration
		ExpVerStatus   *authms.SMSVerificationStatus
	}{
		{
			Desc:   "Valid args",
			ExpErr: false,
			SMSer:  &SMSerMock{ExpErr: nil},
			Tokener: &TokenerMock{
				ExpErr:   nil,
				ExpToken: "some-token",
				ExpJwt:   &jwt.Token{Raw: "some-token"},
			},
			SecureRandomer: &SecureRandomerMock{
				ExpErr: nil, ExpString: "1234",
			},
			CodeValidity: 5 * time.Minute,
			SMSFormat:    "verification code is %s",
		},
		{
			Desc:   "Missing SMSer",
			ExpErr: true,
			Tokener: &TokenerMock{
				ExpErr:   nil,
				ExpToken: "some-token",
				ExpJwt:   &jwt.Token{Raw: "some-token"},
			},
			SMSer: nil,
			SecureRandomer: &SecureRandomerMock{
				ExpErr: nil, ExpString: "1234",
			},
			CodeValidity: 5 * time.Minute,
			SMSFormat:    "verification code is %s",
		},
		{
			Desc:    "Missing Tokener",
			ExpErr:  true,
			Tokener: nil,
			SMSer:   &SMSerMock{ExpErr: nil},
			SecureRandomer: &SecureRandomerMock{
				ExpErr: nil, ExpString: "1234",
			},
			CodeValidity: 5 * time.Minute,
			SMSFormat:    "verification code is %s",
		},
		{
			Desc:   "Missing secure Randomer",
			ExpErr: true,
			SMSer:  &SMSerMock{ExpErr: nil},
			Tokener: &TokenerMock{
				ExpErr:   nil,
				ExpToken: "some-token",
				ExpJwt:   &jwt.Token{Raw: "some-token"},
			},
			SecureRandomer: nil,
			CodeValidity:   5 * time.Minute,
			SMSFormat:      "verification code is %s",
		},
		{
			Desc:   "Bad SMS format code",
			ExpErr: true,
			SMSer:  &SMSerMock{ExpErr: nil},
			Tokener: &TokenerMock{
				ExpErr:   nil,
				ExpToken: "some-token",
				ExpJwt:   &jwt.Token{Raw: "some-token"},
			},
			SecureRandomer: &SecureRandomerMock{
				ExpErr: nil, ExpString: "1234",
			},
			CodeValidity: 5 * time.Minute,
			SMSFormat:    "verification code is",
		},
		{
			Desc:   "Empty SMS format code",
			ExpErr: true,
			SMSer:  &SMSerMock{ExpErr: nil},
			Tokener: &TokenerMock{
				ExpErr:   nil,
				ExpToken: "some-token",
				ExpJwt:   &jwt.Token{Raw: "some-token"},
			},
			SecureRandomer: &SecureRandomerMock{
				ExpErr: nil, ExpString: "1234",
			},
			CodeValidity: 5 * time.Minute,
			SMSFormat:    "",
		},
		{
			Desc:   "Bad code validity",
			ExpErr: true,
			SMSer:  &SMSerMock{ExpErr: nil},
			Tokener: &TokenerMock{
				ExpErr:   nil,
				ExpToken: "some-token",
				ExpJwt:   &jwt.Token{Raw: "some-token"},
			},
			SecureRandomer: &SecureRandomerMock{
				ExpErr: nil, ExpString: "1234",
			},
			CodeValidity: 59 * time.Second,
			SMSFormat:    "verification code is %s",
		},
		{
			Desc:   "missing code validity",
			ExpErr: true,
			SMSer:  &SMSerMock{ExpErr: nil},
			Tokener: &TokenerMock{
				ExpErr:   nil,
				ExpToken: "some-token",
				ExpJwt:   &jwt.Token{Raw: "some-token"},
			},
			SecureRandomer: &SecureRandomerMock{
				ExpErr: nil, ExpString: "1234",
			},
			SMSFormat: "verification code is %s",
		},
	}
	for _, tc := range tcs {
		v, err := verification.New(tc.SMSFormat, tc.CodeValidity, tc.SMSer, tc.SecureRandomer, tc.Tokener)
		if tc.ExpErr {
			if err == nil {
				fmt.Println(tc.CodeValidity)
				t.Errorf("%s - expected an error but got nil", tc.Desc)
			}
			continue
		} else if err != nil {
			t.Errorf("%s - verification.New(): %s", tc.Desc, err)
			continue
		}
		if v == nil {
			t.Errorf("%s - found nil verifier", tc.Desc)
		}
	}
}

func TestVerifier_SendSMSCode(t *testing.T) {
	validDeps := validDependencies()
	validDeps.ExpVerStatus = &authms.SMSVerificationStatus{
		Token:     "some-token",
		ExpiresAt: time.Now().Add(5 * time.Minute).Format(time.RFC3339),
		Verified:  false,
	}
	tcs := []VerifierTestCase{
		validDeps,
		{
			Desc:   "SMSer not implemented",
			ExpErr: true, ExpNotImpl: true,
			SMSer: &SMSerMock{ExpErr: errors.NewNotImplemented()},
			Tokener: &TokenerMock{
				ExpErr:   nil,
				ExpToken: "some-token",
				ExpJwt:   &jwt.Token{Raw: "some-token"},
			},
			SecureRandomer: &SecureRandomerMock{
				ExpErr: nil, ExpString: "1234",
			},
			CodeValidity: 5 * time.Minute,
			SMSFormat:    "verification code is %s",
			ExpVerStatus: &authms.SMSVerificationStatus{
				Token:     "some-token",
				ExpiresAt: time.Now().Add(5 * time.Minute).Format(time.RFC3339),
				Verified:  false,
			},
		},
	}
	for _, tc := range tcs {
		v, err := verification.New(tc.SMSFormat, tc.CodeValidity, tc.SMSer, tc.SecureRandomer, tc.Tokener)
		if err != nil {
			t.Errorf("%s - verification.New(): %s", tc.Desc, err)
			continue
		}
		stts, err := v.SendSMSCode(tc.Phone)
		if tc.ExpErr {
			if err == nil {
				t.Errorf("%s - expected an error but got nil",
					tc.Desc)
			}
			continue
		} else if err != nil {
			t.Errorf("%s - verifier.SendSMSCode(): %s", tc.Desc, err)
			continue
		}
		if !tc.SMSer.SMSCalled {
			t.Errorf("%s - SMSer.SMS() was not called", tc.Desc)
		}
		if !tc.Tokener.GenerateWithClaimsCalled {
			t.Errorf("%s - Tokener.GenerateWithClaims() was not called", tc.Desc)
		}
		if stts.Token != tc.ExpVerStatus.Token {
			t.Errorf("%s - expected token %v but got %v", tc.Desc,
				tc.ExpVerStatus.Token, stts.Token)
		}
		if stts.Verified != tc.ExpVerStatus.Verified {
			t.Errorf("%s - expected verified status %v but got %v",
				tc.Desc, tc.ExpVerStatus.Verified, stts.Verified)
		}
	}
}

func TestVerifier_VerifySMSCode(t *testing.T) {
	validDeps := validDependencies()
	validDeps.ExpVerStatus = &authms.SMSVerificationStatus{Verified: true}
	validDeps.Tokener.ExpJwt = &jwt.Token{Valid: true}
	validDeps.Tokener.ExpCode = "123"
	validDeps.CodeReq = &authms.SMSVerificationCodeRequest{Code: "123"}
	tcs := []VerifierTestCase{
		validDeps,
	}
	for _, tc := range tcs {
		v, err := verification.New(tc.SMSFormat, tc.CodeValidity, tc.SMSer, tc.SecureRandomer, tc.Tokener)
		if err != nil {
			t.Errorf("%s - verification.New(): %s", tc.Desc, err)
			continue
		}
		stts, err := v.VerifySMSCode(tc.CodeReq)
		if tc.ExpErr {
			if err == nil {
				t.Errorf("%s - expected an error but got nil",
					tc.Desc)
			}
			continue
		} else if err != nil {
			t.Errorf("%s - verifier.VerifySMSCode(): %s", tc.Desc, err)
			continue
		}
		if !tc.Tokener.ValidateClaimsCalled {
			t.Errorf("%s - Tokener.ValidateClaimsCalled() was not"+
				" called", tc.Desc)
		}
		if stts.Verified != tc.ExpVerStatus.Verified {
			t.Errorf("%s - expected verified status %v but got %v",
				tc.Desc, tc.ExpVerStatus.Verified, stts.Verified)
		}
	}
}
