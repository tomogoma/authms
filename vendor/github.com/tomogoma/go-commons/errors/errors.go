package errors

import "fmt"

// Error implements the Error interface and helps distinguish whether an error
// is a client error or an auth error.
type Error struct {
	IsAuthErr           bool
	IsUnauthorizedErr   bool
	IsForbiddenErr      bool
	IsClErr             bool
	IsNotFoundErr       bool
	IsNotImplementedErr bool
	Message             string
}

// Error returns the error message of the error (without the distinguishing flags
// such as client error).
func (e Error) Error() string {
	return e.Message
}

// Client returns true if this is a client error.
func (e Error) Client() bool {
	return e.IsClErr
}

// NotImplemented returns true if the functionality requested is not implemented.
func (e Error) NotImplemented() bool {
	return e.IsNotImplementedErr
}

// Auth returns true if this is an auth error.
func (e Error) Auth() bool {
	return e.IsAuthErr || e.IsUnauthorizedErr || e.IsForbiddenErr
}

// Unauthorized returns true if this is an Unauthorized error.
func (e Error) Unauthorized() bool {
	return e.IsUnauthorizedErr
}

// Forbidden returns true if this is a Forbidden error.
func (e Error) Forbidden() bool {
	return e.IsForbiddenErr
}

// NotFound returns true if this is error denotes that a resource
// being fetched was not found.
func (e Error) NotFound() bool {
	return e.IsNotFoundErr
}

// New creates a new error.
func New(message string) Error {
	return Error{Message: message, IsClErr: false}
}

// Newf creates a new error with fmt.Printf formatting.
func Newf(format string, a ...interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return New(message)
}

// NewClient creates a new client error.
func NewClient(message string) Error {
	return Error{Message: message, IsClErr: true}
}

// NewClientf creates a new client error with fmt.Printf style formatting.
func NewClientf(format string, a ...interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return NewClient(message)
}

// NewNotImplemented creates a new not implemented error.
func NewNotImplemented() Error {
	return Error{IsNotImplementedErr: true}
}

// NewNotImplementedf creates a new not implemented error with fmt.Printf
// style formatting.
func NewNotImplementedf(format string, a ...interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return Error{Message: message, IsNotImplementedErr: true}
}

// NewAuth creates a new auth error. It is not specific to the type of auth
// error. Use NewForbidden(string) or NewUnauthorized(string) to establish
// a more specific Auth error.
func NewAuth(message string) Error {
	return Error{Message: message, IsAuthErr: true}
}

// NewForbidden creates a new forbidden auth error a la 403 (http.StatusForbidden) error.
// This will also resolve as an Auth error.
func NewForbidden(message string) Error {
	return Error{Message: message, IsAuthErr: true, IsForbiddenErr: true}
}

// NewUnauthorized creates a new unauthorized auth error a la 401 (http.StatusUnauthorized) error.
// This will also resolve as an Auth error.
func NewUnauthorized(message string) Error {
	return Error{Message: message, IsAuthErr: true, IsUnauthorizedErr: true}
}

// NewAuthf creates a new auth error with fmt.Printf style formatting.
func NewAuthf(format string, a ...interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return NewAuth(message)
}

// NewForbiddenf creates a new forbidden auth error with fmt.Printf style formatting.
// This will also resolve as an Auth error.
func NewForbiddenf(format string, a ...interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return NewForbidden(message)
}

// NewUnauthorizedf creates a new unauthorized auth error with fmt.Printf style formatting.
// This will also resolve as an Auth error.
func NewUnauthorizedf(format string, a ...interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return NewUnauthorized(message)
}

// NewNotFound creates a new not found error.
func NewNotFound(message string) Error {
	return Error{Message: message, IsNotFoundErr: true}
}

// NewNotFoundf creates a new not found error with fmt.Printf style formatting.
func NewNotFoundf(format string, a ...interface{}) Error {
	message := fmt.Sprintf(format, a...)
	return NewNotFound(message)
}

// ClErrCheck is a helper struct that can be embedded in a custom struct to
// give the custom struct the extra method IsClientError(err error). e.g:
//  type Custom struct {
//      ...
//      errors.ClErrCheck
//  }
type ClErrCheck struct {
}

// IsClientError returns true if the supplied error is a client error, false otherwise.
func (c *ClErrCheck) IsClientError(err error) bool {
	errC, ok := err.(Error)
	return ok && errC.Client()
}

// NotImplErrCheck is a helper struct that can be embedded in a custom struct to
// give the custom struct the extra method IsNotFoundError(err error). e.g:
//  type Custom struct {
//      ...
//      errors.NotImplErrCheck
//  }
type NotImplErrCheck struct {
}

// IsNotImplementedError returns true if the supplied error is a client error, false otherwise.
func (c *NotImplErrCheck) IsNotImplementedError(err error) bool {
	errC, ok := err.(Error)
	return ok && errC.NotImplemented()
}

// AuthErrCheck is a helper struct that can be embedded in a custom struct to
// give the custom struct the extra methods IsAuthError(err error),
// IsForbiddenError(err error) and IsUnauthorizedErr(err error) . e.g:
//  type Custom struct {
//      ...
//      errors.AuthErrCheck
//  }
type AuthErrCheck struct {
}

// IsAuthError returns true if the supplied error is an
// authentication/authorization error, false otherwise.
func (c *AuthErrCheck) IsAuthError(err error) bool {
	errC, ok := err.(Error)
	return ok && errC.Auth()
}

// IsAuthError returns true if the supplied error is an
// authentication/authorization error, false otherwise.
func (c *AuthErrCheck) IsForbiddenError(err error) bool {
	errC, ok := err.(Error)
	return ok && errC.Forbidden()
}

// IsAuthError returns true if the supplied error is an
// authentication/authorization error, false otherwise.
func (c *AuthErrCheck) IsUnauthorizedError(err error) bool {
	errC, ok := err.(Error)
	return ok && errC.Unauthorized()
}

// NotFoundErrCheck is a helper struct that can be embedded in a custom struct to
// give the custom struct the extra method IsNotFoundError(err error). e.g:
//  type Custom struct {
//      ...
//      errors.NotFoundErrCheck
//  }
type NotFoundErrCheck struct {
}

// IsNotFoundError returns true if the supplied error is an not found error, false otherwise.
func (c *NotFoundErrCheck) IsNotFoundError(err error) bool {
	errC, ok := err.(Error)
	return ok && errC.NotFound()
}
