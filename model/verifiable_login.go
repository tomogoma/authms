package model

import "time"

type VerifLogin struct {
	ID         string
	UserID     string
	Address    string
	Verified   bool
	CreateDate time.Time
	UpdateDate time.Time
}
