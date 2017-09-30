package model

import (
	"github.com/dgrijalva/jwt-go"
)

type Claim struct {
	UsrID  string
	Groups []Group
	jwt.StandardClaims
}
