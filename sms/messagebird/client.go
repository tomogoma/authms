package messagebird

import (
	"fmt"

	"github.com/messagebird/go-rest-api"
)

type Client struct {
	cl      *messagebird.Client
	accName string
}

type SMSError struct {
	sms *messagebird.Message
	err error
}

func (e SMSError) Error() string {
	var msg string
	if e.err != nil {
		msg = e.err.Error()
	} else {
		msg = "message bird API returned an error"
	}
	if e.sms == nil || len(e.sms.Errors) == 0 {
		return msg
	}
	msgs := "["
	for _, err := range e.sms.Errors {
		msgs = fmt.Sprintf("%s %+v", msgs, err)
	}
	msgs = msgs + " ]"
	return fmt.Sprintf("%s: %s", msg, msgs)
}

func NewClient(accName, apiKey string) (*Client, error) {
	client := messagebird.New(apiKey)
	return &Client{accName: accName, cl: client}, nil
}

func (at *Client) SMS(toPhone, message string) error {
	msg, err := at.cl.NewMessage(at.accName, []string{toPhone}, message, nil)
	if err != nil {
		return SMSError{sms: msg, err: err}
	}
	return nil
}
