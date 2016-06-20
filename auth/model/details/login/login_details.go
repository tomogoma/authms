package login

import (
	"time"

	"errors"
)

var ErrorBadUserID = errors.New("User ID provided was not good")

type LoginDetails struct {
	id         int
	userID     int
	ipAddress  string
	date       time.Time
	forService string
	referral   string
}

func (ld *LoginDetails) ID() int            { return ld.id }
func (ld *LoginDetails) UserID() int        { return ld.userID }
func (ld *LoginDetails) IPAddress() string  { return ld.ipAddress }
func (ld *LoginDetails) Date() time.Time    { return ld.date }
func (ld *LoginDetails) ForService() string { return ld.forService }
func (ld *LoginDetails) Referral() string   { return ld.referral }

func New(uID int, date time.Time, ip, forSrvc, ref string) (*LoginDetails, error) {

	if uID == 0 {
		return nil, ErrorBadUserID
	}

	return &LoginDetails{
		userID:     uID,
		ipAddress:  ip,
		date:       date,
		forService: forSrvc,
		referral:   ref,
	}, nil
}
