package smtp

import (
	"strings"

	"html/template"

	"github.com/tomogoma/authms/model"
)

type email struct {
	Recipients string
	Subject    string
	Body       template.HTML
}

const emailTmplt = `To: {{.Recipients}}
Subject: {{.Subject}}
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";
Content-Transfer-Encoding: 8bit;
{{.Body}}
`

func newEmail(r model.SendMail) email {
	m := email{
		Subject: r.Subject,
		Body:    r.Body,
	}
	for _, addr := range r.ToEmails {
		m.Recipients = m.Recipients + addr + ","
	}
	m.Recipients = strings.TrimSuffix(m.Recipients, ",")
	return m
}
