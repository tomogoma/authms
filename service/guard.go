package service

import (
	"time"

	"strings"

	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/generator"
	"github.com/tomogoma/go-commons/errors"
)

type APIKeyStore interface {
	IsNotFoundError(error) bool
	InsertAPIKey(userID, key string) (*APIKey, error)
	APIKeysByUserID(userID string, offset, count int64) ([]APIKey, error)
}

type KeyGenerator interface {
	SecureRandomBytes(length int) ([]byte, error)
}

type APIKey struct {
	ID         string
	UserID     string
	APIKey     string
	CreateDate time.Time
	UpdateDate time.Time
}

type Guard struct {
	errors.ClErrCheck
	errors.AuthErrCheck
	db        APIKeyStore
	gen       KeyGenerator
	masterKey string
}

type Option func(*Guard)

func WithMasterKey(key string) Option {
	return func(g *Guard) {
		g.masterKey = key
	}
}

func WithKeyGenerator(kg KeyGenerator) Option {
	return func(g *Guard) {
		g.gen = kg
	}
}

var invalidAPIKeyErrorf = "invalid API key (%s) for %s"

func NewGuard(db APIKeyStore, opts ...Option) (*Guard, error) {
	if db == nil {
		return nil, errors.New("APIKeyStore was nil")
	}
	g := &Guard{db: db}
	for _, f := range opts {
		f(g)
	}
	if g.gen == nil {
		var err error
		g.gen, err = generator.NewRandom(generator.AlphaNumericChars)
		if err != nil {
			return nil, errors.Newf("creating random number generator")
		}
	}
	return g, nil
}

func (s *Guard) APIKeyValid(key string) (string, error) {
	if key != "" && key == s.masterKey {
		return "master", nil
	}
	pair := strings.SplitN(key, ".", 2)
	if len(pair) < 2 || pair[0] == "" || pair[1] == "" {
		return "", errors.NewUnauthorizedf("invalid API key %+v", pair)
	}
	userID := pair[0]
	keyStr := pair[1]
	dbKeys, err := s.db.APIKeysByUserID(userID, 0, 10)
	if err != nil {
		if s.db.IsNotFoundError(err) {
			return userID, errors.NewForbiddenf(invalidAPIKeyErrorf, keyStr, userID)
		}
		return userID, errors.Newf("get API Key: %v", err)
	}
	for _, dbKey := range dbKeys {
		if dbKey.APIKey == keyStr {
			return userID, nil
		}
	}
	return userID, errors.NewForbiddenf(invalidAPIKeyErrorf, keyStr, userID)
}

func (s *Guard) NewAPIKey(userID string) (*APIKey, error) {
	if userID == "" {
		return nil, errors.NewClient("userID was empty")
	}
	key, err := s.gen.SecureRandomBytes(config.APIKeyLength)
	if err != nil {
		return nil, errors.Newf("generate key: %v", err)
	}
	k, err := s.db.InsertAPIKey(userID, string(key))
	if err != nil {
		return nil, errors.Newf("store key: %v", err)
	}
	return k, nil
}
