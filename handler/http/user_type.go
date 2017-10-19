package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
 * @api {NULL} UserType
 * @apiName UserType
 * @apiGroup Objects
 *
 * @apiSuccess {String} ID Unique ID of the userType (can be cast to long Integer).
 * @apiSuccess {String} name Unique name of the user type.
 * @apiSuccess {String} created ISO8601 date the user type was created.
 * @apiSuccess {String} lastUpdated ISO8601 date the user type was last updated.
 */
type UserType struct {
	ID         string `json:"ID,omitempty"`
	Name       string `json:"name,omitempty"`
	CreateDate string `json:"created,omitempty"`
	UpdateDate string `json:"lastUpdated,omitempty"`
}

func NewUserType(ut model.UserType) *UserType {
	if !ut.HasValue() {
		return nil
	}
	return &UserType{
		ID:         ut.ID,
		Name:       ut.Name,
		CreateDate: ut.CreateDate.Format(config.TimeFormat),
		UpdateDate: ut.UpdateDate.Format(config.TimeFormat),
	}
}
