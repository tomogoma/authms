package model

import "testing"

func TestNormalizeUsername(t *testing.T) {
	tt := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "all lower case", input: "johndoe", expect: "johndoe"},
		{name: "With upper case", input: "JohnDoe", expect: "johndoe"},
		{name: "Need normalize1", input: "žůžo", expect: "zuzo"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeValidUsername(tc.input)
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if actual != tc.expect {
				t.Errorf("Expected %s, got %s", tc.expect, actual)
			}
		})
	}
}

func TestNormalizeValidEmail(t *testing.T) {
	tt := []struct {
		name   string
		input  string
		expect string
		expErr bool
	}{
		{name: "all lower case", input: "test@mailinator.com", expect: "test@mailinator.com"},
		{name: "With upper case", input: "Test@Mailinator.com", expect: "test@mailinator.com"},
		{name: "With + in username", input: "Test+one@Mailinator.com", expErr: true},
		{name: "With . in username", input: "Test.one@Mailinator.com", expErr: true},
		{name: "Bad format", input: "ç$€§/az@mailinator.com", expErr: true},
		{name: "Bad domain", input: "email@x-unkown-domain.com", expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeValidEmail(tc.input)
			if tc.expErr {
				if err == nil {
					t.Errorf("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if actual != tc.expect {
				t.Errorf("Expected %s, got %s", tc.expect, actual)
			}
		})
	}
}
