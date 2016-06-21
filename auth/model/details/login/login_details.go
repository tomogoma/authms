package login

import (
	"time"

	"errors"
)

var ErrorBadUserID = errors.New("User ID provided was not good")

type LoginDetails interface {
	ID() int
	UserID() int
	IPAddress() string
	Date() time.Time
	ForService() string
	Referral() string
}

type loginDetails struct {
	id         int
	userID     int
	ipAddress  string
	date       time.Time
	forService string
	referral   string
}

func (ld *loginDetails) ID() int            { return ld.id }
func (ld *loginDetails) UserID() int        { return ld.userID }
func (ld *loginDetails) IPAddress() string  { return ld.ipAddress }
func (ld *loginDetails) Date() time.Time    { return ld.date }
func (ld *loginDetails) ForService() string { return ld.forService }
func (ld *loginDetails) Referral() string   { return ld.referral }

func New(uID int, date time.Time, ip, forSrvc, ref string) (*loginDetails, error) {

	if uID == 0 {
		return nil, ErrorBadUserID
	}

	return &loginDetails{
		userID:     uID,
		ipAddress:  ip,
		date:       date,
		forService: forSrvc,
		referral:   ref,
	}, nil
}
