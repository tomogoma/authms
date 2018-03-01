package model

import "time"

type VerifLogin struct {
	ID         string
	UserID     string
	Address    string
	Verified   bool
	OTPStatus  DBTStatus
	CreateDate time.Time
	UpdateDate time.Time
}

func (vl VerifLogin) HasValue() bool {
	return vl.ID != ""
}
