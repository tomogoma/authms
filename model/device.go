package model

import "time"

type Device struct {
	ID         string
	UserID     string
	DeviceID   string
	CreateDate time.Time
	UpdateDate time.Time
}
