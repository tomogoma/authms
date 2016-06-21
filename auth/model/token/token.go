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

var ErrorEmptyDevID = errors.New("devID cannot be empty")

type Token interface {
	ID() int
	UserID() int
	DevID() string
	Token() string
	Issued() time.Time
	Expiry() time.Time
}

type token struct {
	id     int
	userID int
	devID  string
	token  string
	issued time.Time
	expiry time.Time
}

type ExpiryType int

const (
	ShortExpType = iota
	MedExpType
	LongExpType
)

func (t *token) ID() int           { return t.id }
func (t *token) UserID() int       { return t.userID }
func (t *token) DevID() string     { return t.devID }
func (t *token) Token() string     { return t.token }
func (t *token) Issued() time.Time { return t.issued }
func (t *token) Expiry() time.Time { return t.expiry }

func New(usrID int, devID string, expType ExpiryType) (*token, error) {

	if usrID < 1 {
		return nil, ErrorBadUserID
	}

	if devID == "" {
		return nil, ErrorEmptyDevID
	}

	tknStr := uuid.NewV4().String()
	issued := time.Now()
	expiry := issued.Add(shortDuration)

	switch expType {
	case MedExpType:
		expiry = issued.Add(mediumDuration)
	case LongExpType:
		expiry = issued.Add(longDuration)
	}

	return &token{
		userID: usrID,
		devID:  devID,
		token:  tknStr,
		issued: issued,
		expiry: expiry,
	}, nil
}
