package verification_test

import (
	"testing"
	"github.com/tomogoma/authms/auth/phone/verification"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/authms/proto/authms"
	"time"
)

type SMSerMock struct {
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
}

func (t *TokenerMock) GenerateWithClaims(claims jwt.Claims) (string, error) {
	t.GenerateWithClaimsCalled = true
	return t.ExpToken, t.ExpErr
}
func (t *TokenerMock) ValidateClaims(token string, claims jwt.Claims) (*jwt.Token, error) {
	t.ValidateClaimsCalled = true
	return t.ExpJwt, t.ExpErr
}

type VerifierTestCase struct {
	Desc         string
	ExpErr       bool
	Phone        string
	CodeReq      *authms.SMSVerificationCodeRequest
	SMSer        *SMSerMock
	Tokener      *TokenerMock
	ExpVerStatus *authms.SMSVerificationStatus
}

func TestNew(t *testing.T) {
	tcs := []VerifierTestCase{
		{
			Desc: "Valid args",
			ExpErr: false,
			SMSer: &SMSerMock{ExpErr: nil},
			Tokener: &TokenerMock{
				ExpErr: nil,
				ExpToken: "some-token",
				ExpJwt: &jwt.Token{Raw: "some-token"},
			},
		},
		{
			Desc: "Missing SMSer",
			ExpErr: false,
			Tokener: &TokenerMock{
				ExpErr: nil,
				ExpToken: "some-token",
				ExpJwt: &jwt.Token{Raw: "some-token"},
			},
		},
		{
			Desc: "Missing Tokener",
			ExpErr: false,
			SMSer: &SMSerMock{ExpErr: nil},
		},
	}
	for _, tc := range tcs {
		v, err := verification.New(tc.SMSer, tc.Tokener)
		if tc.ExpErr {
			if err == nil {
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
	tcs := []VerifierTestCase{
		{
			Desc: "successful send",
			ExpErr: false,
			Phone: "01234",
			ExpVerStatus: &authms.SMSVerificationStatus{
				Token: "some-token",
				Retries: int32(3), ExpiresAt: time.Now().Add(5 * time.Minute).Format(time.RFC3339),
				IsBlocked: false, BlockedUntil:"", Verified:false,
			},
			SMSer: &SMSerMock{ExpErr: nil},
			Tokener: &TokenerMock{
				ExpErr: nil,
				ExpToken: "some-token",
				ExpJwt: &jwt.Token{Raw: "some-token"},
			},
		},
	}
	for _, tc := range tcs {
		v, err := verification.New(tc.SMSer, tc.Tokener)
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
		if stts.Retries != tc.ExpVerStatus.Retries {
			t.Errorf("%s - expected retries %v but got %v", tc.Desc,
				tc.ExpVerStatus.Retries, stts.Retries)
		}
		if stts.IsBlocked != tc.ExpVerStatus.IsBlocked {
			t.Errorf("%s - expected blocked status %v but got %v",
				tc.Desc, tc.ExpVerStatus.IsBlocked, stts.IsBlocked)
		}
		if stts.Verified != tc.ExpVerStatus.Verified {
			t.Errorf("%s - expected verified status %v but got %v",
				tc.Desc, tc.ExpVerStatus.Verified, stts.Verified)
		}
	}
}

func TestVerifier_VerifySMSCode(t *testing.T) {
	tcs := []VerifierTestCase{
		{
			Desc: "Valid token/code combo",
			ExpErr: false,
			CodeReq:  &authms.SMSVerificationCodeRequest{Token:"some-token", Code:"some-code"},
			ExpVerStatus: &authms.SMSVerificationStatus{Verified:true},
			SMSer: &SMSerMock{ExpErr: nil},
			Tokener: &TokenerMock{
				ExpErr: nil,
				ExpToken: "some-token",
				ExpJwt: &jwt.Token{Raw: "some-token"},
			},
		},
	}
	for _, tc := range tcs {
		v, err := verification.New(tc.SMSer, tc.Tokener)
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
			t.Errorf("%s - Tokener.ValidateClaimsCalled() was not" +
				" called", tc.Desc)
		}
		if stts.Verified != tc.ExpVerStatus.Verified {
			t.Errorf("%s - expected verified status %v but got %v",
				tc.Desc, tc.ExpVerStatus.Verified, stts.Verified)
		}
	}
}
