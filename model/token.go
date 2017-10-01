package model

import "time"

type Token struct {
	ID         string
	UserID     string
	Phone      string
	Email      string
	IsUsed     bool
	Token      []byte
	IssueDate  time.Time
	ExpiryDate time.Time
}
