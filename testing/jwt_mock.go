package testing

import "github.com/dgrijalva/jwt-go"

type JWTMock struct {
	ExpGenJWT    string
	ExpGenJWTErr error
	ExpValTkn    *jwt.Token
	ExpVarErr    error
}

func (j *JWTMock) Generate(claims jwt.Claims) (string, error) {
	return j.ExpGenJWT, j.ExpGenJWTErr
}
func (j *JWTMock) Validate(JWT string, claims jwt.Claims) (*jwt.Token, error) {
	return j.ExpValTkn, j.ExpVarErr
}
