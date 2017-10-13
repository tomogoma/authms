package model

import "time"

type UserType struct {
	ID         string
	Name       string
	CreateDate time.Time
	UpdateDate time.Time
}

func (ut UserType) HasValue() bool {
	return ut.ID != ""
}
