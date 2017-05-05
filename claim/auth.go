package claim

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Auth struct {
	UsrID int64
	DevID string
	jwt.StandardClaims
}

func NewAuth(UsrID int64, DevID string, validity time.Duration) Auth {
	now := time.Now()
	expiry := now.Add(validity)
	return Auth{
		UsrID: UsrID,
		DevID: DevID,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: expiry.Unix(),
			Issuer:    "authms",
		},
	}
}
