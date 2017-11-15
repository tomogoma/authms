package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
 * @apiDefine VerifLogin
 *
 * @apiSuccess {String} ID Unique ID of the verifiable login (can be cast to long Integer).
 * @apiSuccess {String} userID ID for user who owns this verifiable login.
 * @apiSuccess {String} value The unique verifiable login string value.
 * @apiSuccess {Boolean} verified True if this login is verified, false otherwise.
 * @apiSuccess {String} created ISO8601 date the verifiable login was created.
 * @apiSuccess {String} lastUpdated ISO8601 date the verifiable login was last updated.
 */

/**
 * @api {NULL} VerifLogin Verifiable Login
 * @apiName VerifLogin
 * @apiVersion 0.1.0
 * @apiGroup Objects
 *
 * @apiUse VerifLogin
 */
type VerifLogin struct {
	ID         string `json:"ID,omitempty"`
	UserID     string `json:"userID,omitempty"`
	Address    string `json:"value,omitempty"`
	Verified   bool   `json:"verified"`
	CreateDate string `json:"created,omitempty"`
	UpdateDate string `json:"lastUpdated,omitempty"`
}

func NewVerifLogin(vl *model.VerifLogin) *VerifLogin {
	if vl == nil || !vl.HasValue() {
		return nil
	}
	return &VerifLogin{
		ID:         vl.ID,
		UserID:     vl.UserID,
		Address:    vl.Address,
		Verified:   vl.Verified,
		CreateDate: vl.CreateDate.Format(config.TimeFormat),
		UpdateDate: vl.UpdateDate.Format(config.TimeFormat),
	}
}
