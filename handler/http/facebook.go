package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

type Facebook struct {
	ID         string `json:"ID,omitempty"`
	UserID     string `json:"userID,omitempty"`
	FacebookID string `json:"facebookID,omitempty"`
	Verified   bool   `json:"verified,omitempty"`
	CreateDate string `json:"createDate,omitempty"`
	UpdateDate string `json:"updateDate,omitempty"`
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
