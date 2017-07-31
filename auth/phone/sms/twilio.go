package sms

import (
	"encoding/json"
	"github.com/tomogoma/go-commons/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	accountsURL  = "https://api.twilio.com/2010-04-01/Accounts/"
	messagesPath = "Messages.json"
	httpMethod   = "POST"
	formKeyTo    = "To"
	formKeyFrom  = "From"
	formKeyBody  = "Body"
)

type TwConfig interface {
	TwilioID() string
	TwilioTokenKeyFile() string
	TwilioSenderPhone() string
}

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

type Twilio struct {
	token  string
	config TwConfig
}

func NewTwilio(c TwConfig) (*Twilio, error) {
	if c == nil {
		return nil, errors.New("Config was nil")
	}
	token, err := readToken(c.TwilioTokenKeyFile())
	if err != nil {
		return nil, err
	}
	return &Twilio{config: c, token: token}, nil
}

func (s *Twilio) SMS(toPhone, message string) error {
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
	URL.Path = path.Join(URL.Path, s.config.TwilioID(), messagesPath)
	form := url.Values{}
	form.Add(formKeyTo, toPhone)
	form.Add(formKeyFrom, s.config.TwilioSenderPhone())
	form.Add(formKeyBody, message)
	req, err := http.NewRequest(httpMethod, URL.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return errors.Newf("unable to create message request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.config.TwilioID(), s.token)
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

func readToken(tokenKeyFile string) (string, error) {
	fb, err := ioutil.ReadFile(tokenKeyFile)
	if err != nil {
		return "", errors.Newf("error opening twilio token file: %v", err)
	}
	return string(fb), nil
}
