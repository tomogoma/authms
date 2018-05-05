// token package imnplements common methods for handling tokens.
package token

import (
	"io/ioutil"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/go-typed-errors"
)

type KeyFiler interface {
	// KeyFile returns the path to a key file.
	KeyFile() string
}

// JWT Handler handles JWT tokens with the help of the github.com/dgrijalva/jwt-go library.
// Call NewJWTHandler() to get the correct instance of this struct.
type JWTHandler struct {
	hs256key []byte
	errors.AuthErrCheck
}

// NewJWTHandler constructs a JWTHandler struct for handling JWT tokens.
func NewJWTHandler(c KeyFiler) (*JWTHandler, error) {
	if c == nil {
		return nil, errors.New("token KeyFiler was nil")
	}
	prk, err := ioutil.ReadFile(c.KeyFile())
	if err != nil {
		return nil, errors.Newf("error getting key from '%s': %s",
			c.KeyFile(), err)
	}
	if len(prk) == 0 {
		return nil, errors.New("Key file was empty")
	}
	return &JWTHandler{hs256key: prk}, nil
}

// Generate generates a JWT token based on provided claims using jwt.SigningMethodHS256
// to sign the token string. jwt.StandardClaims is a good starting point for a claims
// struct and can be extended to implement the jwt.Claims interface.
func (g *JWTHandler) Generate(claims jwt.Claims) (string, error) {
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tkn.SignedString(g.hs256key)
}

// Validate validates a token that was generated using jwt.SigningMethodHS256 to
// sign the token string.
//
// the error - err - returned will evaluate (*JWTHandler).IsAuthError(err) to true if
// the token is invalid, otherwise any other error is returned if the parameters
// are invalid e.g. nil claims parameter.
//
// The claims parameter should be a pointer to the struct to which the (valid)
// token is unmarshalled into.
//
// The returned *jwt.Token provides standard information on the token.
func (g *JWTHandler) Validate(token string, claims jwt.Claims) (*jwt.Token, error) {
	if claims == nil {
		return nil, errors.New("claims were nil")
	}
	tkn, err := jwt.ParseWithClaims(
		token, claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.NewAuth("invalid token")
			}
			return g.hs256key, nil
		})
	if err != nil {
		return nil, errors.NewAuth("invalid token")
	}
	return tkn, nil
}
