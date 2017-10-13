package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

type UserType struct {
	ID         string `json:"ID,omitempty"`
	Name       string `json:"name,omitempty"`
	CreateDate string `json:"createDate,omitempty"`
	UpdateDate string `json:"updateDate,omitempty"`
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
