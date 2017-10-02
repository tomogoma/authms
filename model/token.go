package model

import "time"

type DBToken struct {
	ID         string
	UserID     string
	Phone      string
	Email      string
	IsUsed     bool
	Token      []byte
	IssueDate  time.Time
	ExpiryDate time.Time
}
