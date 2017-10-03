package model_test

import (
	"testing"

	"github.com/tomogoma/authms/mocks"
	"github.com/tomogoma/authms/model"
)

func TestNewAuthentication(t *testing.T) {
	tt := []struct {
		name   string
		db     *mocks.DBMock
		guard  *mocks.GuardMock
		jwter  *mocks.JWTMock
		opts   []model.Option
		expErr bool
	}{
		{
			name:   "min valid deps",
			db:     &mocks.DBMock{},
			guard:  &mocks.GuardMock{},
			jwter:  &mocks.JWTMock{},
			expErr: false,
		},
		{
			name:   "nil db",
			db:     nil,
			guard:  &mocks.GuardMock{},
			jwter:  &mocks.JWTMock{},
			expErr: true,
		},
		{
			name:   "nil guard",
			db:     &mocks.DBMock{},
			guard:  nil,
			jwter:  &mocks.JWTMock{},
			expErr: true,
		},
		{
			name:   "nil jwter",
			db:     &mocks.DBMock{},
			guard:  &mocks.GuardMock{},
			jwter:  nil,
			expErr: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			a, err := model.NewAuthentication(tc.db, tc.guard, tc.jwter, tc.opts...)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got an error: %v", err)
			}
			if a == nil {
				t.Fatalf("Got nil *Authentication")
			}
		})
	}
}
