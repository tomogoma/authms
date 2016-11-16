package oauth

import (
	"errors"
	"fmt"
)

type Config struct {
	FacebookSecret string `json:"facebookSecret,omitempty"`
	FacebookID     int    `json:"facebookID,omitempty"`
}

var ErrorEmptyFacebookSecret = errors.New("facebook secret was empty")

func (c Config) Validate() error {
	fmt.Printf("%+v\n", c)
	if c.FacebookSecret == "" {
		return ErrorEmptyFacebookSecret
	}
	return nil
}
