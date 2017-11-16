package model

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/authms/config"
)

type JWTClaim struct {
	UsrID          string
	Group         Group
	jwt.StandardClaims
}

func newJWTClaim(usrID string, group Group) *JWTClaim {
	issue := time.Now()
	expiry := issue.Add(tokenValidity)
	return &JWTClaim{
		UsrID:          usrID,
		Group:         group,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issue.Unix(),
			ExpiresAt: expiry.Unix(),
			Issuer:    config.CanonicalName(),
		},
	}
}
