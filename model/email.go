package model

import "html/template"

type SendMail struct {
	ToEmails []string
	Subject  string
	Body     template.HTML
}
