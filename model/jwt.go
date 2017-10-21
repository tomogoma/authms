package model

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/authms/config"
)

type JWTClaim struct {
	UsrID          string
	StrongestGroup *Group
	Groups         []Group
	jwt.StandardClaims
}

func newJWTClaim(usrID string, groups []Group) *JWTClaim {
	var strongestGrp *Group
	for _, grp := range groups {
		if strongestGrp == nil || grp.AccessLevel < strongestGrp.AccessLevel {
			strongestGrp = &grp
		}
	}
	issue := time.Now()
	expiry := issue.Add(tokenValidity)
	return &JWTClaim{
		UsrID:          usrID,
		Groups:         groups,
		StrongestGroup: strongestGrp,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issue.Unix(),
			ExpiresAt: expiry.Unix(),
			Issuer:    config.CanonicalName(),
		},
	}
}
