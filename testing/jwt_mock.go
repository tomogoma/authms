package testing

import "github.com/dgrijalva/jwt-go"

type JWTMock struct {
	ExpGenJWTErr error
	ExpValTkn    *jwt.Token
	ExpVarErr    error
}

func (j *JWTMock) Generate(claims jwt.Claims) (string, error) {
	return "a.good.jwt", j.ExpGenJWTErr
}
func (j *JWTMock) Validate(JWT string, claims jwt.Claims) (*jwt.Token, error) {
	return j.ExpValTkn, j.ExpVarErr
}
