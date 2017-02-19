package oauth

import (
	"errors"

	"github.com/tomogoma/authms/auth/oauth/facebook"
	"github.com/tomogoma/authms/auth/oauth/response"
	"io/ioutil"
	"fmt"
)

const (
	AppFacebook = "facebook"
)

type OAuthClient interface {
	ValidateToken(token string) (response.OAuth, error)
}

type OAuth struct {
	fb OAuthClient
}

var ErrorUnsupportedApp = errors.New("the app provided is not supported")

func New(c Config) (*OAuth, error) {
	secretBytes, err := ioutil.ReadFile(c.FacebookSecretFileLoc)
	if err != nil {
		return nil, fmt.Errorf("error reading facebook secret from" +
			" file: %s", err)
	}
	fb, err := facebook.New(c.FacebookID, string(secretBytes))
	if err != nil {
		return nil, err
	}
	return &OAuth{fb: fb}, nil
}

func (o *OAuth) ValidateToken(appName, token string) (response.OAuth, error) {
	switch appName {
	case AppFacebook:
		return o.fb.ValidateToken(token)
	default:
		return nil, ErrorUnsupportedApp
	}
}
