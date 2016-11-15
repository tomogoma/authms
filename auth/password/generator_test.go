package password_test

import (
	"testing"

	"fmt"

	"github.com/tomogoma/authms/auth/password"
)

func TestNewGenerator(t *testing.T) {
	type testcase struct {
		desc    string
		charSet string
		expErr  error
	}

	edgeTooLongCharSet := ""
	for i := 0; i < 257; i++ {
		edgeTooLongCharSet += "a"
	}

	testcases := []testcase{
		{
			desc:    "good characterset",
			charSet: password.AllChars,
			expErr:  nil,
		},
		{
			desc:    "empty characterset",
			charSet: "",
			expErr:  password.ErrorBadCharSet,
		},
		{
			desc:    "too long a characterset",
			charSet: edgeTooLongCharSet,
			expErr:  password.ErrorBadCharSet,
		},
	}

	for _, tc := range testcases {
		g, err := password.NewGenerator(tc.charSet)
		if err != tc.expErr {
			t.Errorf("%s: Expected error %v but got %v",
				tc.desc, tc.expErr, err)
			continue
		}
		if tc.expErr == nil && g == nil {
			t.Errorf("%s: got nil generator", tc.desc)
		}
	}
}

func TestSecureRandomBytes(t *testing.T) {
	g, err := password.NewGenerator(password.AllChars)
	if err != nil {
		t.Fatalf("NewGenerator: %s", err)
	}
	size := 36
	p, err := g.SecureRandomString(size)
	if err != nil {
		t.Fatalf("SecureRandomString: %s", err)
	}
	fmt.Printf("Password generated %s\n", string(p))
	if len(p) != size {
		t.Errorf("Expected password of length %d but got %d", size, len(p))
	}
}
