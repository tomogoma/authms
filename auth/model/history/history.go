package history

import (
	"time"

	"errors"
)

const (
	LoginAccess           = 1 << iota
	RegistrationAccess    = 1 << iota
	TokenValidationAccess = 1 << iota
)

var ErrorBadUserID = errors.New("User ID provided was not good")
var ErrorBadAccessMethod = errors.New("Access Method was unknown")

type History struct {
	id           int
	userID       int
	ipAddress    string
	date         time.Time
	accessMethod int
	forService   string
	referral     string
	successful   bool
}

func (ld *History) ID() int            { return ld.id }
func (ld *History) UserID() int        { return ld.userID }
func (ld *History) IPAddress() string  { return ld.ipAddress }
func (ld *History) Date() time.Time    { return ld.date }
func (ld *History) AccessMethod() int  { return ld.accessMethod }
func (ld *History) ForService() string { return ld.forService }
func (ld *History) Referral() string   { return ld.referral }
func (ld *History) Successful() bool   { return ld.successful }

func (h *History) Validate() error {

	if h.userID == 0 {
		return ErrorBadUserID
	}

	if err := validateAcM(h.accessMethod); err != nil {
		return err
	}

	return nil
}

func New(uID, acM int, successful bool, date time.Time, ip, forSrvc, ref string) (*History, error) {

	h := &History{
		userID:       uID,
		accessMethod: acM,
		successful:   successful,
		ipAddress:    ip,
		date:         date,
		forService:   forSrvc,
		referral:     ref,
	}

	if err := h.Validate(); err != nil {
		return nil, err
	}

	return h, nil
}

func validateAcM(acM int) error {

	if acM != TokenValidationAccess && acM != LoginAccess && acM != RegistrationAccess {
		return ErrorBadAccessMethod
	}
	return nil
}
