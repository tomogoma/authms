package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
 * @apiDefine OTPStatus
 *
 * @apiSuccess {String} obfuscatedAddress Obfuscated address to which OTP was sent.
 * @apiSuccess {String} expiresAt ISO8601 expiry date of OTP.
 */

/**
 * @api {NULL} OTPStatus OTP Status
 * @apiName OTPStatus
 * @apiVersion 0.1.0
 * @apiGroup Objects
 *
 * @apiUse OTPStatus
 */
type DBTStatus struct {
	ObfuscatedAddress string `json:"obfuscatedAddress,omitempty"`
	ExpiresAt         string `json:"expiresAt,omitempty"`
}

func NewDBTStatus(dbtS *model.DBTStatus) *DBTStatus {
	if dbtS == nil || !dbtS.HasValue() {
		return nil
	}
	return &DBTStatus{
		ObfuscatedAddress: dbtS.ObfuscatedAddress,
		ExpiresAt:         dbtS.ExpiresAt.Format(config.TimeFormat),
	}
}
