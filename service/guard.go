package service

import (
	"time"

	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/generator"
	"github.com/tomogoma/go-commons/errors"
	"golang.org/x/crypto/bcrypt"
	"github.com/tomogoma/authms/model"
)

type APIKeyStore interface {
	IsNotFoundError(error) bool
	GetAPIKeys(userID string) ([][]byte, error)
	SaveAPIKey(userID string, key []byte) (*model.APIKey, error)
}

type KeyGenerator interface {
	SecureRandomBytes(length int) ([]byte, error)
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

var invalidAPIKeyErrorf = "invalid api key ('%s') for '%s'"

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

func (s *Guard) APIKeyValid(userID, key string) error {
	if userID == "" || key == "" {
		return errors.NewUnauthorizedf(invalidAPIKeyErrorf, key, userID)
	}
	if key == s.masterKey {
		return nil
	}
	keysHB, err := s.db.GetAPIKeys(userID)
	if err != nil {
		if s.db.IsNotFoundError(err) {
			return errors.NewForbiddenf(invalidAPIKeyErrorf, key, userID)
		}
		return errors.Newf("get API Key: %v", err)
	}
	for _, keyHB := range keysHB {
		err = bcrypt.CompareHashAndPassword(keyHB, []byte(key))
		if err == nil {
			return nil
		}
	}
	return errors.NewForbiddenf(invalidAPIKeyErrorf, key, userID)
}

func (s *Guard) NewAPIKey(userID string) (*model.APIKey, error) {
	if userID == "" {
		return nil, errors.NewClient("userID was empty")
	}
	key, err := s.gen.SecureRandomBytes(config.APIKeyLength)
	if err != nil {
		return nil, errors.Newf("generate key: %v", err)
	}
	keyH, err := bcrypt.GenerateFromPassword(key, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Newf("hash key for storage: %v", err)
	}
	k, err := s.db.SaveAPIKey(userID, keyH)
	if err != nil {
		return nil, errors.Newf("store key: %v", err)
	}
	k.APIKey = string(key)
	return k, nil
}
