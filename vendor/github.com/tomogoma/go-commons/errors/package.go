// errors package is an implementation of go's Error interface useful in API setups.
// errors trys to distinguish between different causes of error e.g. client error,
// authentication/authorization error ...or general error (which would generally
// imply that the error is not the client's fault ~ in an API setup).
package errors
