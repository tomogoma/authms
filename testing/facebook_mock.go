package testing

import errors "github.com/tomogoma/go-typed-errors"

type FacebookMock struct {
	errors.AuthErrCheck
	ExpValTknErr error
}

func (f *FacebookMock) ValidateToken(token string) (string, error) {
	if f.ExpValTknErr != nil {
		return "", f.ExpValTknErr
	}
	return currentID(), f.ExpValTknErr
}
