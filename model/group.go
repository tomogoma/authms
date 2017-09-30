package model

import "time"

type Group struct {
	ID          string
	Name        string
	AccessLevel int
	CreateDate  time.Time
	UpdateDate  time.Time
}
