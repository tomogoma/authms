package model

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Claim struct {
	UsrID int64
	DevID string
	jwt.StandardClaims
}

func NewClaim(UsrID int64, DevID string, validity time.Duration) Claim {
	now := time.Now()
	expiry := now.Add(validity)
	return Claim{
		UsrID: UsrID,
		DevID: DevID,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: expiry.Unix(),
			Issuer:    "authms",
		},
	}
}
