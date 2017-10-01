package smtp

import (
	"strings"

	"html/template"

	"github.com/tomogoma/authms/model"
)

type email struct {
	Recipients string
	Subject    string
	Greeting   template.HTML
	Body       template.HTML
	Signature  template.HTML
	Footer     template.HTML
}

const emailTmplt = `To: {{.Recipients}}
Subject: {{.Subject}}
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";
Content-Transfer-Encoding: 8bit;
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html>
<p>
	{{.Greeting}},
</p>
<div>
	{{.Body}}
	<p>{{.Signature}}</p>
</div>
<div>
	{{.Footer}}
</div>
</html>
`

func newEmail(r model.SendMail) email {
	m := email{
		Subject:   r.Subject,
		Greeting:  "Hello",
		Body:      r.Body,
		Signature: r.Signature,
		Footer:    r.Footer,
	}
	if r.RecipientName != "" {
		m.Greeting = "Dear " + r.RecipientName
	}
	for _, addr := range r.ToEmails {
		m.Recipients = m.Recipients + addr + ","
	}
	m.Recipients = strings.TrimSuffix(m.Recipients, ",")
	return m
}
