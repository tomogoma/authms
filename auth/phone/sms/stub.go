package sms

import "github.com/tomogoma/go-commons/errors"

type Stub struct {
	errors.NotImplErrCheck
}

func (s Stub) SMS(toPhone, message string) error {
	return errors.NewNotImplemented()
}
