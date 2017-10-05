package model

import "time"

type Group struct {
	ID          string
	Name        string
	AccessLevel float32
	CreateDate  time.Time
	UpdateDate  time.Time
}
