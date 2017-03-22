package errors

import (
	"github.com/tomogoma/go-commons/errors"
	"fmt"
)

type Error struct {
	IsAuthErr bool
	Message   string
}

type IsClientErrorer struct {
	errors.ClErrCheck
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) Auth() bool {
	return e.IsAuthErr
}

func NewAuth(message string) error {
	return Error{Message: message, IsAuthErr: true}
}

func NewAuthf(format string, a... interface{}) error {
	message := fmt.Sprintf(format, a...)
	return NewAuth(message)
}

func New(message string) error {
	return errors.New(message)
}

func Newf(format string, a... interface{}) error {
	return errors.Newf(format, a)
}

func NewClient(message string) error {
	return errors.NewClient(message)
}

func NewClientf(format string, a... interface{}) error {
	return errors.NewClientf(format, a)
}


