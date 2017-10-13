package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

type Username struct {
	ID         string `json:"ID,omitempty"`
	UserID     string `json:"userID,omitempty"`
	Value      string `json:"value,omitempty"`
	CreateDate string `json:"updateDate,omitempty"`
	UpdateDate string `json:"createDate,omitempty"`
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
