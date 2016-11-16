package oauth

import "errors"

type Config struct {
	FacebookSecret string `json:"facebookSecret,omitempty"`
	FacebookID     int    `json:"facebookID,omitempty"`
}

var ErrorEmptyFacebookSecret = errors.New("facebook secret was empty")
var ErrorBadFacebookID = errors.New("facebook id was invalid")

func (c Config) Validate() error {
	if c.FacebookSecret == "" {
		return ErrorEmptyFacebookSecret
	}
	if c.FacebookID < 1 {
		return ErrorBadFacebookID
	}
	return nil
}
