package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

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
	CreateDate string      `json:"createDate,omitempty"`
	UpdateDate string      `json:"updateDate,omitempty"`
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
