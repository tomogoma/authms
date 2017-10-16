package facebook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	errors "github.com/tomogoma/go-typed-errors"
)

const (
	fbURL           = "https://graph.facebook.com/debug_token"
	fbAppTokenKey   = "access_token"
	fbInputTokenKey = "input_token"
)

type FacebookOAuth struct {
	errors.AuthErrCheck
	appID     int64
	appSecret string
}

var ErrorEmptyAppSecret = errors.New("facebook app secret was empty")
var ErrorEmptyAppID = errors.New("facebook app ID was empty")

func New(appID int64, appSecret string) (*FacebookOAuth, error) {
	if appSecret == "" {
		return nil, ErrorEmptyAppSecret
	}
	if appID == 0 {
		return nil, ErrorEmptyAppID
	}
	return &FacebookOAuth{appID: appID, appSecret: appSecret}, nil
}

func (f *FacebookOAuth) ValidateToken(token string) (string, error) {
	URL, err := url.Parse(fmt.Sprintf("%s?%s=%s&%s=%d|%s", fbURL,
		fbInputTokenKey, token, fbAppTokenKey, f.appID, f.appSecret))
	if err != nil {
		return "", fmt.Errorf("error parsing facebook url: %s", err)
	}
	r, err := http.Get(URL.String())
	if err != nil {
		return "", fmt.Errorf("error communicating with facebook: %s", err)
	}
	defer r.Body.Close()
	if r.StatusCode >= 400 {
		return "", fmt.Errorf("error communicating with facebook (%d): %s",
			r.StatusCode, r.Status)
	}
	rb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("error reading facebook response: %s", err)
	}
	resp := new(Response)
	err = json.Unmarshal(rb, resp)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling facebook response: %s", err)
	}
	if !resp.Valid {
		return "", errors.NewAuth("OAuth token invalid")
	}
	return resp.UsrID, nil
}
