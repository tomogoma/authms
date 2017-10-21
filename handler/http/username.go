package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
 * @api {NULL} Username Username
 * @apiName Username
 * @apiVersion 0.1.0
 * @apiGroup Objects
 *
 * @apiSuccess {String} ID Unique ID of the username (can be cast to long Integer).
 * @apiSuccess {String} userID ID for user who owns this Username.
 * @apiSuccess {String} value The unique username string value.
 * @apiSuccess {String} created ISO8601 date the username was created.
 * @apiSuccess {String} lastUpdated ISO8601 date the username was last updated.
 */
type Username struct {
	ID         string `json:"ID,omitempty"`
	UserID     string `json:"userID,omitempty"`
	Value      string `json:"value,omitempty"`
	CreateDate string `json:"created,omitempty"`
	UpdateDate string `json:"lastUpdated,omitempty"`
}

func NewUserName(un model.Username) *Username {
	if !un.HasValue() {
		return nil
	}
	return &Username{
		ID:         un.ID,
		UserID:     un.UserID,
		Value:      un.Value,
		CreateDate: un.CreateDate.Format(config.TimeFormat),
		UpdateDate: un.UpdateDate.Format(config.TimeFormat),
	}
}
