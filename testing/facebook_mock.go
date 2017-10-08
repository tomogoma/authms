package testing

import "github.com/tomogoma/go-commons/errors"

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
