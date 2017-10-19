package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
 * @api {NULL} User
 * @apiName User
 * @apiGroup Objects
 *
 * @apiSuccess {String} ID Unique ID of the user (can be cast to long Integer).
 * @apiSuccess {String} JWT JSON Web Token for accessing services.
 * @use {Object} type See <a href="#api-Objects-UserType">UserType</a>.
 * @apiSuccess {Object} username See <a href="#api-Objects-Username">Username</a>.
 * @apiSuccess {Object} username See <a href="#api-Objects-Username">Username</a>.
 * @apiSuccess {Object} phone See <a href="#api-Objects-VerifLogin">VerifLogin</a>.
 * @apiSuccess {Object} email See <a href="#api-Objects-VerifLogin">VerifLogin</a>.
 * @apiSuccess {Object} facebook See <a href="#api-Objects-FacebookID">FacebookID</a>.
 * @apiSuccess {Object} group See <a href="#api-Objects-Group">Group</a>.
 * @apiSuccess {Object} device See <a href="#api-Objects-Device">Device</a>.
 * @apiSuccess {String} created Date the user was created.
 * @apiSuccess {String} lastUpdated date the user was last updated.
 */
type User struct {
	ID         string      `json:"ID,omitempty"`
	JWT        string      `json:"JWT,omitempty"`
	Type       *UserType   `json:"type,omitempty"`
	UserName   *Username   `json:"username,omitempty"`
	Phone      *VerifLogin `json:"phone,omitempty"`
	Email      *VerifLogin `json:"email,omitempty"`
	Facebook   *Facebook   `json:"facebook,omitempty"`
	Groups     []Group     `json:"groups,omitempty"`
	Devices    []Device    `json:"devices,omitempty"`
	CreateDate string      `json:"created,omitempty"`
	UpdateDate string      `json:"lastUpdated,omitempty"`
}

func NewUser(user *model.User) *User {
	if user == nil {
		return nil
	}
	return &User{
		ID:         user.ID,
		JWT:        user.JWT,
		Type:       NewUserType(user.Type),
		UserName:   NewUserName(user.UserName),
		Phone:      NewVerifLogin(&user.Phone),
		Email:      NewVerifLogin(&user.Email),
		Facebook:   NewFacebook(user.Facebook),
		Groups:     NewGroups(user.Groups),
		Devices:    NewDevices(user.Devices),
		CreateDate: user.CreateDate.Format(config.TimeFormat),
		UpdateDate: user.UpdateDate.Format(config.TimeFormat),
	}
}
