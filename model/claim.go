package model

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/authms/config"
)

type Claim struct {
	UsrID  string
	Groups []Group
	jwt.StandardClaims
}

func newClaim(usrID string, groups []Group) *Claim {
	issue := time.Now()
	expiry := issue.Add(tokenValidity)
	return &Claim{
		UsrID:  usrID,
		Groups: groups,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issue.Unix(),
			ExpiresAt: expiry.Unix(),
			Issuer:    config.CanonicalName,
		},
	}
}
