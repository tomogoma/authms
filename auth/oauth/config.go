package oauth

import "errors"

type Config struct {
	FacebookSecretFileLoc string `json:"facebookSecretFileLoc,omitempty" yaml:"facebookSecretFileLoc"`
	FacebookID            int    `json:"facebookID,omitempty" yaml:"facebookID"`
}

var ErrorEmptyFacebookSecretFileLoc = errors.New("facebook secret file location was empty")
var ErrorBadFacebookID = errors.New("facebook id was invalid")

func (c Config) Validate() error {
	if c.FacebookSecretFileLoc == "" {
		return ErrorEmptyFacebookSecretFileLoc
	}
	if c.FacebookID < 1 {
		return ErrorBadFacebookID
	}
	return nil
}
