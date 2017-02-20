package sms

import (
	"net/http"
	"github.com/tomogoma/go-commons/errors"
	"path"
	"io/ioutil"
	"net/url"
	"strings"
)

const (
	accountsURL = "https://api.twilio.com/2010-04-01/Accounts/"
	messagesPath = "Messages.json"
	httpMethod = "POST"
	formKeyTo = "To"
	formKeyFrom = "From"
	formKeyBody = "Body"
	testMessage = "this is a test message"
)

type Config interface {
	TwilioID() string
	TwilioTokenKeyFile() string
	TwilioSenderPhone() string
	TwilioTestNumber() string
}

type SMS struct {
	token  string
	config Config
}

func New(c Config) (*SMS, error) {
	if c == nil {
		return nil, errors.New("Config was nil")
	}
	token, err := readToken(c.TwilioTokenKeyFile())
	if err != nil {
		return nil, err
	}
	sms := &SMS{config:c, token:token}
	if err := sms.SMS(c.TwilioTestNumber(), testMessage); err != nil {
		return nil, errors.Newf("Unable to access twilio, probably" +
			" SMS config values are invalid?: %s", err)
	}
	return sms, nil
}

func (s *SMS) SMS(toPhone, message string) error {
	client := &http.Client{}
	URL, err := url.Parse(accountsURL)
	if err != nil {
		return errors.Newf("problem with the twilio URL (%v)" +
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