package model

import "time"

type Facebook struct {
	ID            string
	UserID        string
	FacebookID    string
	FacebookToken string
	Verified      bool
	CreateDate    time.Time
	UpdateDate    time.Time
}

func (fb Facebook) HasValue() bool {
	return fb.ID != ""
}
