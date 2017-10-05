package service_test

import (
	"testing"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/service"
	"github.com/tomogoma/go-commons/errors"
)

type APIKeyStoreMock struct {
	errors.NotFoundErrCheck
	expKeys   []model.APIKey
	expAPIKey *model.APIKey
	expErr    error
}

type KeyGeneratorMock struct {
	expErr error
	expKey []byte
}

func (kg *KeyGeneratorMock) SecureRandomBytes(length int) ([]byte, error) {
	return kg.expKey, kg.expErr
}

func (db *APIKeyStoreMock) APIKeysByUserID(userID string, offset, count int64) ([]model.APIKey, error) {
	return db.expKeys, db.expErr
}
func (db *APIKeyStoreMock) InsertAPIKey(userID, key string) (*model.APIKey, error) {
	return db.expAPIKey, db.expErr
}

func TestNewGuard(t *testing.T) {
	tt := []struct {
		name      string
		db        service.APIKeyStore
		kg        service.KeyGenerator
		masterKey string
		expErr    bool
	}{
		{
			name:      "all deps provided",
			db:        &APIKeyStoreMock{},
			kg:        &KeyGeneratorMock{},
			masterKey: "a-master-key",
			expErr:    false,
		},
		{
			name:      "implicit deps",
			db:        &APIKeyStoreMock{},
			kg:        nil,
			masterKey: "",
			expErr:    false,
		},
		{
			name:      "nil db",
			db:        nil,
			kg:        &KeyGeneratorMock{},
			masterKey: "a-master-key",
			expErr:    true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g, err := service.NewGuard(
				tc.db,
				service.WithKeyGenerator(tc.kg),
				service.WithMasterKey(tc.masterKey),
			)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected error got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("service.NewGuard(): %v", err)
			}
			if g == nil {
				t.Fatalf("service.NewGuard returned nil *service.Guard")
			}
		})
	}
}

func TestGuard_NewAPIKey(t *testing.T) {
	validKey := "some-api-key"
	tt := []struct {
		name     string
		userID   string
		kg       *KeyGeneratorMock
		db       *APIKeyStoreMock
		expErr   bool
		expClErr bool
	}{
		{
			name:   "valid",
			userID: "johndoe",
			kg:     &KeyGeneratorMock{expKey: []byte(validKey)},
			db:     &APIKeyStoreMock{expAPIKey: &model.APIKey{UserID: "johndoe", ID: "apiid", APIKey: validKey}},
			expErr: false,
		},
		{
			name:     "missing userID",
			userID:   "",
			kg:       &KeyGeneratorMock{expKey: []byte(validKey)},
			db:       &APIKeyStoreMock{expAPIKey: &model.APIKey{UserID: "johndoe", ID: "apiid", APIKey: validKey}},
			expErr:   true,
			expClErr: true,
		},
		{
			name:   "key gen report error",
			userID: "johndoe",
			kg:     &KeyGeneratorMock{expErr: errors.Newf("an error")},
			db:     &APIKeyStoreMock{expAPIKey: &model.APIKey{UserID: "johndoe", ID: "apiid", APIKey: validKey}},
			expErr: true,
		},
		{
			name:   "db report error",
			userID: "johndoe",
			kg:     &KeyGeneratorMock{expKey: []byte(validKey)},
			db:     &APIKeyStoreMock{expErr: errors.Newf("whoops, an error")},
			expErr: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := newGuard(t, "", tc.kg, tc.db)
			ak, err := g.NewAPIKey(tc.userID)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Error: %v", err)
				}
				if tc.expClErr != g.IsClientError(err) {
					t.Errorf("Expect service.Guard#IsClientError %t, got %t",
						tc.expClErr, g.IsClientError(err))
				}
				return
			}
			if ak == nil {
				t.Fatalf("yielded nil *service.APIKey")
			}
			if ak.APIKey != validKey {
				t.Errorf("API Key mismatch: expect '%s' got '%s'",
					validKey, ak.APIKey)
			}
		})
	}
}

func TestGuard_APIKeyValid(t *testing.T) {
	validKey := "some-api-key"
	tt := []struct {
		name            string
		userID          string
		key             string
		db              *APIKeyStoreMock
		masterKey       string
		expErr          bool
		expForbidden    bool
		expUnauthorized bool
	}{
		{
			name:   "valid (db)",
			userID: "johndoe",
			key:    validKey,
			db: &APIKeyStoreMock{expKeys: []model.APIKey{
				{APIKey: "first-api-key"},
				{APIKey: validKey},
				{APIKey: "last-api-key"},
			}},
			expErr: false,
		},
		{
			name:      "valid (master)",
			userID:    "johndoe",
			key:       "the-master-key",
			masterKey: "the-master-key",
			db: &APIKeyStoreMock{expKeys: []model.APIKey{
				{APIKey: "first-api-key"},
				{APIKey: validKey},
				{APIKey: "last-api-key"},
			}},
			expErr: false,
		},
		{
			name:   "missing userID",
			userID: "",
			key:    validKey,
			db: &APIKeyStoreMock{expKeys: []model.APIKey{
				{APIKey: "first-api-key"},
				{APIKey: validKey},
				{APIKey: "last-api-key"},
			}},
			expErr:          true,
			expForbidden:    false,
			expUnauthorized: true,
		},
		{
			name:   "missing key",
			userID: "johndoe",
			key:    "",
			db: &APIKeyStoreMock{expKeys: []model.APIKey{
				{APIKey: "first-api-key"},
				{APIKey: validKey},
				{APIKey: "last-api-key"},
			}},
			expErr:          true,
			expForbidden:    false,
			expUnauthorized: true,
		},
		{
			name:   "invalid key",
			userID: "johndoe",
			key:    "some-invalid-key",
			db: &APIKeyStoreMock{expKeys: []model.APIKey{
				{APIKey: "first-api-key"},
				{APIKey: validKey},
				{APIKey: "last-api-key"},
			}},
			expErr:          true,
			expForbidden:    true,
			expUnauthorized: false,
		},
		{
			name:            "none found",
			userID:          "johndoe",
			key:             validKey,
			db:              &APIKeyStoreMock{expErr: errors.NewNotFound("no keys for johndoe")},
			expErr:          true,
			expForbidden:    true,
			expUnauthorized: false,
		},
		{
			name:   "db report error",
			userID: "johndoe",
			key:    "some-invalid-key",
			db:     &APIKeyStoreMock{expErr: errors.New("some errors")},
			expErr: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := newGuard(t, tc.masterKey, &KeyGeneratorMock{}, tc.db)
			err := g.APIKeyValid(tc.userID, tc.key)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error, got nil")
				}
				if tc.expUnauthorized != g.IsUnauthorizedError(err) {
					t.Errorf("Expect service.Guard#IsUnauthorizedError %t, got %t",
						tc.expUnauthorized, g.IsUnauthorizedError(err))
				}
				if tc.expForbidden != g.IsForbiddenError(err) {
					t.Errorf("Expect service.Guard#IsForbiddenError %t, got %t",
						tc.expForbidden, g.IsForbiddenError(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected nil error, got %v", err)
			}
		})
	}
}

func newGuard(t *testing.T, master string, kg service.KeyGenerator, db service.APIKeyStore) *service.Guard {
	g, err := service.NewGuard(
		db,
		service.WithKeyGenerator(kg),
		service.WithMasterKey(master),
	)
	if err != nil {
		t.Fatalf("service.NewGuard(): %v", err)
	}
	return g
}
