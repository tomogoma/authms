package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

type DBTStatus struct {
	ObfuscatedAddress string `json:"obfuscatedAddress,omitempty"`
	ExpiresAt         string `json:"expiresAt,omitempty"`
}

func NewDBTStatus(dbtS *model.DBTStatus) *DBTStatus {
	if dbtS == nil {
		return nil
	}
	return &DBTStatus{
		ObfuscatedAddress: dbtS.ObfuscatedAddress,
		ExpiresAt:         dbtS.ExpiresAt.Format(config.TimeFormat),
	}
}
