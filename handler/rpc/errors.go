package rpc

import (
	"github.com/tomogoma/authms/logging"
	"golang.org/x/net/context"
	errors "github.com/tomogoma/go-typed-errors"
	"net/http"
	"fmt"
)

type Error struct {
	Code int
	Message string
}

func (e Error) String() string {
	return fmt.Sprintf(`{"code": %d, "message":"%s"}`, e.Code, e.Message)
}

func (h *UsersHandler) processError(ctx context.Context, err error) error {

	log := ctx.Value(ctxKeyLog).(logging.Logger)

	if h.usersM.IsForbiddenError(err) || h.IsForbiddenError(err) {
		log.Warnf("Forbidden: %v", err)
		return errors.New(Error{Code: http.StatusForbidden, Message: err.Error()})
	}
	if h.usersM.IsAuthError(err) || h.IsAuthError(err){
		log.Warnf("Unauthorized: %v", err)
		return errors.New(Error{Code: http.StatusUnauthorized, Message: err.Error()})
	}

	if h.usersM.IsClientError(err) || h.IsClientError(err) {
		log.Warnf("Bad request: %v", err)
		return errors.New(Error{Code: http.StatusBadRequest, Message: err.Error()})
	}

	if h.usersM.IsNotFoundError(err) || h.IsNotFoundError(err) {
		log.Warnf("Not found: %v", err)
		return errors.New(Error{Code: http.StatusNotFound, Message: err.Error()})
	}

	if h.usersM.IsNotImplementedError(err) || h.IsNotImplementedError(err) {
		log.Warnf("Not implemented entity: %v", err)
		return errors.New(Error{Code: http.StatusNotImplemented, Message: err.Error()})
	}

	log.Errorf("Internal error: %v", err)
	return errors.New(internalErrorMessage)
}