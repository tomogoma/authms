package errors

import (
	"fmt"
)

type Error struct {
	IsAuthErr bool
	IsClErr   bool
	Message   string
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) Auth() bool {
	return e.IsAuthErr
}

func (e Error) Client() bool {
	return e.IsClErr
}

func New(message string) Error {
	return Error{Message:message, IsClErr: false}
}

func Newf(format string, a... interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return New(message)
}

func NewClient(message string) Error {
	return Error{Message: message, IsClErr: true}
}

func NewClientf(format string, a... interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return NewClient(message)
}

func NewAuth(message string) Error {
	return Error{Message: message, IsClErr: true, IsAuthErr: true}
}

func NewAuthf(format string, a... interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return NewAuth(message)
}


