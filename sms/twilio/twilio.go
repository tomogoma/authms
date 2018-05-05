package twilio

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/tomogoma/go-typed-errors"
)

const (
	accountsURL  = "https://api.twilio.com/2010-04-01/Accounts/"
	messagesPath = "Messages.json"
	httpMethod   = "POST"
	formKeyTo    = "To"
	formKeyFrom  = "From"
	formKeyBody  = "Body"
)

const (
	phoneBlockedRcv   = 30004
	unknownPhone      = 30005
	unreachablePhone  = 3006
	invalidPhone      = 21211
	invalidTrialPhone = 14111
	phoneNoSMS        = 21407
)

var rcvErrors = map[int]string{
	phoneBlockedRcv:   "the phone provided is blocked",
	unknownPhone:      "the phone provided is unknown",
	unreachablePhone:  "the phone provided is unreachable",
	invalidPhone:      "the phone provided is invalid",
	invalidTrialPhone: "the phone provided is invalid",
	phoneNoSMS:        "the phone provided cannot receive SMS",
}

type SMSCl struct {
	token       string
	id          string
	senderPhone string
}

func NewSMSCl(id, token, senderPhone string) (*SMSCl, error) {
	if id == "" {
		return nil, errors.New("id was empty")
	}
	if token == "" {
		return nil, errors.New("token was empty")
	}
	if senderPhone == "" {
		return nil, errors.New("senderPhone was empty")
	}
	if id == "" {
		return nil, errors.New("id was nil")
	}
	return &SMSCl{id: id, senderPhone: senderPhone, token: token}, nil
}

func (s *SMSCl) SMS(toPhone, message string) error {
	if toPhone == "" {
		return errors.New("toPhone was empty")
	}
	if message == "" {
		return errors.New("message was empty")
	}
	client := &http.Client{}
	URL, err := url.Parse(accountsURL)
	if err != nil {
		return errors.Newf("problem with the twilio URL (%v)"+
			" contact support", err)
	}
	URL.Path = path.Join(URL.Path, s.id, messagesPath)
	form := url.Values{}
	form.Add(formKeyTo, toPhone)
	form.Add(formKeyFrom, s.senderPhone)
	form.Add(formKeyBody, message)
	req, err := http.NewRequest(httpMethod, URL.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return errors.Newf("unable to create message request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.id, s.token)
	resp, err := client.Do(req)
	if err != nil {
		return errors.Newf("unable to perform request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		type ResponseError struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		rErr := ResponseError{}
		if err := json.Unmarshal(body, &rErr); err == nil {
			if errMsg, ok := rcvErrors[rErr.Code]; ok {
				return errors.NewClient(errMsg)
			}
		}
		return errors.Newf("%s: %s", resp.Status, body)
	}
	return nil
}
