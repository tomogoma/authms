package token

import (
	"errors"

	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	shortDuration  = 8 * time.Hour
	mediumDuration = 720 * time.Hour
	longDuration   = 4320 * time.Hour
)

var ErrorEmptyUserName = errors.New("uName cannot be empty")
var ErrorEmptyDevID = errors.New("devID cannot be empty")

type Token struct {
	id     int
	userID int
	devID  string
	token  string
	issue  time.Time
	expiry time.Time
}

type ExpiryType int

const (
	ShortExpType = iota
	MedExpType
	LongExpType
)

func (t *Token) ID()     { return t.id }
func (t *Token) UserID() { return t.userID }
func (t *Token) DevID()  { return t.devID }
func (t *Token) Token()  { return t.token }
func (t *Token) Issue()  { return t.issue }
func (t *Token) Expiry() { return t.expiry }

func New(uName, devID string, expType ExpiryType) (*Token, error) {

	if uName == "" {
		return nil, ErrorEmptyUserName
	}

	if devID == "" {
		return nil, ErrorEmptyDevID
	}

	token := uuid.NewV4()
	issue := time.Now()
	expiry := issue.Add(shortDuration)

	switch expType {
	case MedExpType:
		expiry = issue.Add(mediumDuration)
	case LongExpType:
		expiry = issue.Add(longDuration)
	}

	return &Token{
		userID: uName,
		devID:  devID,
		token:  token,
		issue:  issue,
		expiry: expiry,
	}, nil
}
