package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
 * @api {NULL} OTPStatus OTP Status
 * @apiName OTPStatus
 * @apiGroup Objects
 *
 * @apiSuccess {String} obfuscatedAddress Obfuscated address to which OTP was sent.
 * @apiSuccess {String} expiresAt ISO8601 expiry date of OTP.
 */
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
