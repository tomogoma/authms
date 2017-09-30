package model

import "time"

type Phone struct {
	ID         string
	UserID     string
	Number     string
	Verified   bool
	CreateDate time.Time
	UpdateDate time.Time
}
