package token

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Generator struct {
	hs256key []byte
}

type Claims struct {
	UserID int    `json:"userID,omitempty"`
	DevID  string `json:"devID,omitempty"`
	jwt.StandardClaims
}

var ErrorSigningMethodTampered = errors.New("Signing method is tampered with")
var ErrorEmptyKeyFile = errors.New("Key file was empty")

func NewGenerator(c Config) (*Generator, error) {
	prk, err := ioutil.ReadFile(c.TknKeyFile)
	if err != nil {
		return nil, fmt.Errorf("error getting key from '%s': %s",
			c.TokenKeyFile, err)
	}
	if len(prk) == 0 {
		return nil, ErrorEmptyKeyFile
	}
	return &Generator{hs256key: prk}, nil
}

func (g *Generator) Generate(usrID int, devID string, expType ExpiryType) (Token, error) {
	issued := time.Now()
	expiry := issued.Add(shortDuration)
	switch expType {
	case MedExpType:
		expiry = issued.Add(mediumDuration)
	case LongExpType:
		expiry = issued.Add(longDuration)
	}
	claims := Claims{
		UserID: usrID,
		DevID:  devID,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issued.Unix(),
			ExpiresAt: expiry.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(g.hs256key)
	if err != nil {
		return nil, fmt.Errorf("Error generating token from key: %s",
			err)
	}
	return New(usrID, devID, tokenStr, issued, expiry)
}

func (g *Generator) Validate(tokenStr string) (Token, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrorSigningMethodTampered
			}
			return g.hs256key, nil
		})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrorInvalidToken
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, ErrorExpiredToken
	}
	issued := time.Unix(claims.IssuedAt, 0)
	expires := time.Unix(claims.ExpiresAt, 0)
	return New(claims.UserID, claims.DevID, tokenStr, issued, expires)
}
