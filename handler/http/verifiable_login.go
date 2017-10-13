package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

type VerifLogin struct {
	ID         string `json:"ID,omitempty"`
	UserID     string `json:"userID,omitempty"`
	Address    string `json:"value,omitempty"`
	Verified   bool   `json:"verified,omitempty"`
	CreateDate string `json:"createDate,omitempty"`
	UpdateDate string `json:"updateDate,omitempty"`
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
