package auth

import (
	"fmt"
	"time"
)

const (
	MinBlackListFailCount = 3
	MinBlackListWindow    = 30 * time.Minute
)

var ErrorBadMaxFailAttempts = fmt.Errorf("MaxFailAttempts cannot be less than %s", MinBlackListFailCount)
var ErrorBadMaxFailAttemptsDuration = fmt.Errorf("MaxFailAttempts cannot be less than %s", MinBlackListWindow)

type Config struct {
	BlackListFailCount int
	BlacklistWindow    time.Duration
}

func (conf Config) Validate() error {

	if conf.BlackListFailCount < MinBlackListFailCount {
		return ErrorBadMaxFailAttempts
	}

	if conf.BlacklistWindow < MinBlackListWindow {
		return ErrorBadMaxFailAttemptsDuration
	}

	return nil
}
