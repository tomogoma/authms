package sms

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tomogoma/go-commons/errors"
)

const (
	atSendURL = "https://api.africastalking.com/restless/send"

	paramUserName = "username"
	paramAPIKey   = "Apikey"
	paramToPhone  = "to"
	paramMessage  = "message"
)

type ATConfig interface {
	Username() string
	APIKey() string
}

type ATOption func(at *AfricasTalking)

type atResponse struct {
	SMSMessageData struct {
		Recipients struct {
			Recipient []struct {
				Status struct {
					Val string `xml:",chardata"`
				} `xml:"status"`
			} `xml:"Recipient"`
		} `xml:"Recipients"`
	} `xml:"SMSMessageData"`
}

type AfricasTalking struct {
	atSendURL string
	userName  string
	apiKey    string
	errors.NotImplErrCheck
}

func ATWithSendURL(URL string) func(at *AfricasTalking) {
	return func(at *AfricasTalking) {
		at.atSendURL = URL
	}
}

func NewAfricasTalking(usrName, APIKey string, opts ...ATOption) (*AfricasTalking, error) {
	if APIKey == "" {
		return nil, errors.New("API key was empty")
	}
	if usrName == "" {
		return nil, errors.New("API UserName was empty")
	}
	at := &AfricasTalking{atSendURL: atSendURL, userName: usrName, apiKey: APIKey}
	for _, opt := range opts {
		opt(at)
	}
	if at.atSendURL == "" {
		return nil, errors.New("Send URL was empty")
	}
	return at, nil
}

func (at *AfricasTalking) SMS(toPhone, message string) error {
	resp, err := at.sendRequest(toPhone, message)
	if err != nil {
		return errors.Newf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Newf("error connecting to API: %s", resp.Status)
	}
	respBody, err := readRespBody(resp.Body)
	if err != nil {
		return err
	}
	recipients := respBody.SMSMessageData.Recipients.Recipient
	if len(recipients) != 1 {
		return errors.Newf("%d recipients were recorded while expecting 1",
			len(recipients))
	}
	if !strings.EqualFold(strings.TrimSpace(recipients[0].Status.Val), "success") {
		return errors.Newf("API reported an error: %v", recipients[0].Status.Val)
	}
	return nil
}

func (at *AfricasTalking) sendRequest(toPhone, message string) (*http.Response, error) {
	if toPhone == "" {
		return nil, errors.Newf("toPhone was empty")
	}
	if message == "" {
		return nil, errors.Newf("message was empty")
	}
	URL, err := url.Parse(at.atSendURL)
	if err != nil {
		return nil, errors.Newf("error parsing configured send URL: %v", err)
	}
	q := URL.Query()
	q.Add(paramUserName, at.userName)
	q.Add(paramAPIKey, at.apiKey)
	q.Add(paramToPhone, toPhone)
	q.Add(paramMessage, message)
	URL.RawQuery = q.Encode()
	return http.Get(URL.String())
}

func readRespBody(resp io.Reader) (atResponse, error) {
	respBody, err := ioutil.ReadAll(resp)
	if err != nil {
		return atResponse{}, errors.Newf("error reading response body: %v", err)
	}
	respStruct := atResponse{}
	if err := xml.Unmarshal(respBody, &respStruct); err != nil {
		return atResponse{}, errors.Newf("error unmarshalling response body: %v", err)
	}
	return respStruct, nil
}
