package model

import "html/template"

type Email struct {
	Address  string
	Verified bool
}

type SendMail struct {
	Subject       string
	ToEmails      []string
	RecipientName template.HTML
	Body          template.HTML
	Signature     template.HTML
	Footer        template.HTML
}
