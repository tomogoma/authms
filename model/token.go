package model

import "time"

type PhoneToken struct {
	ID         string
	UserID     string
	Phone      string
	IsUsed     bool
	IssueDate  time.Time
	ExpiryDate time.Time
}

type EmailToken struct {
	ID         string
	UserID     string
	Email      string
	IsUsed     bool
	IssueDate  time.Time
	ExpiryDate time.Time
}
