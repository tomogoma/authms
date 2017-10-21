package model

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/authms/config"
)

type JWTClaim struct {
	UsrID  string
	Groups []Group
	jwt.StandardClaims
}

func newJWTClaim(usrID string, groups []Group) *JWTClaim {
	issue := time.Now()
	expiry := issue.Add(tokenValidity)
	return &JWTClaim{
		UsrID:  usrID,
		Groups: groups,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issue.Unix(),
			ExpiresAt: expiry.Unix(),
			Issuer:    config.CanonicalName(),
		},
	}
}
