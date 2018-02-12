// token package imnplements common methods for handling tokens.
package jwt

import (
	"reflect"

	"github.com/dgrijalva/jwt-go"
	errors "github.com/tomogoma/go-typed-errors"
)

// JWT Handler handles JWT tokens with the help of the github.com/dgrijalva/jwt-go library.
// Call NewHandler() to get the correct instance of this struct.
type Handler struct {
	hs256key []byte
	errors.AuthErrCheck
}

// NewHandler constructs a Handler struct for handling JWT tokens.
func NewHandler(hs256key []byte) (*Handler, error) {
	if len(hs256key) == 0 {
		return nil, errors.New("hs256 key was empty")
	}
	return &Handler{hs256key: hs256key}, nil
}

// Generate generates a JWT token based on provided claims using jwt.SigningMethodHS256
// to sign the token string. jwt.StandardClaims is a good starting point for a claims
// struct and can be extended to implement the jwt.Claims interface.
func (g *Handler) Generate(claims jwt.Claims) (string, error) {
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tkn.SignedString(g.hs256key)
}

// Validate validates a token that was generated using jwt.SigningMethodHS256 to
// sign the token string.
//
// the error - err - returned will evaluate (*Handler).IsAuthError(err) to true if
// the token is invalid, otherwise any other error is returned if the parameters
// are invalid e.g. nil claims parameter.
//
// The claims parameter should be a pointer to the struct to which the (valid)
// token is unmarshalled into.
//
// The returned *jwt.Token provides standard information on the token.
func (g *Handler) Validate(token string, cs jwt.Claims) (*jwt.Token, error) {
	if token == "" {
		return nil, errors.NewUnauthorized("token was empty")
	}
	if cs == nil || reflect.ValueOf(cs).IsNil() {
		return nil, errors.New("claims not provided")
	}
	tkn, err := jwt.ParseWithClaims(
		token, cs,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.NewForbidden("invalid token")
			}
			return g.hs256key, nil
		})
	if err != nil {
		return nil, errors.NewForbidden("invalid token")
	}
	return tkn, nil
}
