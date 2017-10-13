package model

import "time"

type Username struct {
	ID         string
	UserID     string
	Value      string
	CreateDate time.Time
	UpdateDate time.Time
}

func (un Username) HasValue() bool {
	return un.ID != ""
}
