package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
@apiDefine User
@apiVersion 0.1.0

@apiSuccess {String} ID				Unique ID of the user
	(can be cast to long Integer).
@apiSuccess {Object} type			The
	<a href="#api-Objects-UserType">UserType</a> of this user.
@apiSuccess {Object} group		The
	<a href="#api-Objects-Group">group</a> the user belongs to.
@apiSuccess {String} created		The date the user was created.
@apiSuccess {String} lastUpdated	date the user was last updated.
@apiSuccess {String} [JWT]			JSON Web Token for accessing services.
	This is only provided during <a href="#api-Auth-Login">Login</a>,
	<a href="#api-Auth-Register">Registration</a>
	and  <a href="#api-Auth-FirstUser">First User Registration</a>.
@apiSuccess {Object} [username]		The user's
	<a href="#api-Objects-Username">username</a> (if this user has one).
@apiSuccess {Object} [phone]		The user's
	<a href="#api-Objects-VerifLogin">phone</a> (if this user has one).
@apiSuccess {Object} [email]		The user's
	<a href="#api-Objects-VerifLogin">email</a> (if this user has one).
@apiSuccess {Object} [facebook] 	The user's
	<a href="#api-Objects-FacebookID">facebook ID</a> (if this user has one).
@apiSuccess {Object} [device]		The
	<a href="#api-Objects-Device">device</a> this user is attached to, if any.
 */

/**
 * @api {JSON} User User
 * @apiName User
 * @apiVersion 0.1.0
 * @apiGroup Objects
 *
 * @apiUse User
 */
type User struct {
	ID         string      `json:"ID,omitempty"`
	JWT        string      `json:"JWT,omitempty"`
	Type       *UserType   `json:"type,omitempty"`
	UserName   *Username   `json:"username,omitempty"`
	Phone      *VerifLogin `json:"phone,omitempty"`
	Email      *VerifLogin `json:"email,omitempty"`
	Facebook   *Facebook   `json:"facebook,omitempty"`
	Group      *Group      `json:"group,omitempty"`
	Devices    []Device    `json:"devices,omitempty"`
	CreateDate string      `json:"created,omitempty"`
	UpdateDate string      `json:"lastUpdated,omitempty"`
}

func NewUser(user *model.User) *User {
	if user == nil || !user.HasValue() {
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
		Group:      NewGroup(user.Group),
		Devices:    NewDevices(user.Devices),
		CreateDate: user.CreateDate.Format(config.TimeFormat),
		UpdateDate: user.UpdateDate.Format(config.TimeFormat),
	}
}

func NewUsers(usrs []model.User) []User {
	if len(usrs) == 0 {
		return nil
	}
	var rtUsrs []User
	for _, usr := range usrs {
		rtUsr := NewUser(&usr)
		if rtUsr == nil {
			continue
		}
		rtUsrs = append(rtUsrs, *rtUsr)
	}
	if len(rtUsrs) == 0 {
		return nil
	}
	return rtUsrs
}
