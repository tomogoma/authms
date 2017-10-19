package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
 * @api {NULL} FacebookID
 * @apiName FacebookID
 * @apiGroup Objects
 *
 * @apiSuccess {String} ID Unique ID of the facebook ID (can be cast to long Integer).
 * @apiSuccess {String} userID ID for user who owns this facebook ID.
 * @apiSuccess {String} facebookID The unique facebook ID string value.
 * @apiSuccess {Boolean} verified True if this login is verified, false otherwise.
 * @apiSuccess {String} created ISO8601 date the facebook ID was inserted.
 * @apiSuccess {String} lastUpdated ISO8601 date the facebook ID value was last updated.
 */
type Facebook struct {
	ID         string `json:"ID,omitempty"`
	UserID     string `json:"userID,omitempty"`
	FacebookID string `json:"facebookID,omitempty"`
	Verified   bool   `json:"verified,omitempty"`
	CreateDate string `json:"created,omitempty"`
	UpdateDate string `json:"lastUpdated,omitempty"`
}

func NewFacebook(fb model.Facebook) *Facebook {
	if !fb.HasValue() {
		return nil
	}
	return &Facebook{
		ID:         fb.ID,
		UserID:     fb.UserID,
		FacebookID: fb.FacebookID,
		Verified:   fb.Verified,
		CreateDate: fb.CreateDate.Format(config.TimeFormat),
		UpdateDate: fb.UpdateDate.Format(config.TimeFormat),
	}
}
