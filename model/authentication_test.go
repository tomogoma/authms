package model_test

import (
	"testing"

	"github.com/tomogoma/authms/model"
	testingH "github.com/tomogoma/authms/testing"
)

func TestNewAuthentication(t *testing.T) {
	tt := []struct {
		name   string
		db     *testingH.DBMock
		guard  *testingH.GuardMock
		jwter  *testingH.JWTMock
		opts   []model.Option
		expErr bool
	}{
		{
			name:   "min valid deps",
			db:     &testingH.DBMock{},
			guard:  &testingH.GuardMock{},
			jwter:  &testingH.JWTMock{},
			expErr: false,
		},
		{
			name:   "nil db",
			db:     nil,
			guard:  &testingH.GuardMock{},
			jwter:  &testingH.JWTMock{},
			expErr: true,
		},
		{
			name:   "nil jwter",
			db:     &testingH.DBMock{},
			guard:  &testingH.GuardMock{},
			jwter:  nil,
			expErr: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			a, err := model.NewAuthentication(tc.db, tc.jwter, tc.opts...)
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
