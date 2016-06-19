package login

import (
	"time"

	"bitbucket.org/alkira/contactsms/kazoo/errors"
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

func (ld *LoginDetails) ID()         { return ld.id }
func (ld *LoginDetails) UserID()     { return ld.userID }
func (ld *LoginDetails) IPAddress()  { return ld.ipAddress }
func (ld *LoginDetails) Date()       { return ld.date }
func (ld *LoginDetails) ForService() { return ld.forService }
func (ld *LoginDetails) Referral()   { return ld.referral }

func New(uID int, date time.Time, ip, forSrvc, ref string) (*LoginDetails, error) {

	if uID == 0 {
		return ErrorBadUserID
	}

	return &LoginDetails{
		userID:     uID,
		ipAddress:  ip,
		date:       date,
		forService: forSrvc,
		referral:   ref,
	}, nil
}
