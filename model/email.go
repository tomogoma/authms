package model

import "html/template"

type Email struct {
	Address  string
	Verified bool
}

type SendMail struct {
	ToEmails []string
	Subject  string
	Body     template.HTML
}
