package model

import "time"

type Device struct {
	ID         string
	UserID     string
	DeviceID   string
	CreateDate time.Time
	UpdateDate time.Time
}

func (d Device) HasValue() bool {
	return d.ID != ""
}