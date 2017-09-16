package auth

import (
	"errors"

	"github.com/tomogoma/authms/auth/oauth/facebook"
	"github.com/tomogoma/authms/auth/oauth/response"
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

func NewOAuth(fbID int64, fbSecret string) (*OAuth, error) {
	fb, err := facebook.New(fbID, fbSecret)
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
