package generator_test

import (
	"testing"

	"fmt"

	"github.com/tomogoma/authms/generator"
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
			charSet: generator.AllChars,
			expErr:  nil,
		},
		{
			desc:    "empty characterset",
			charSet: "",
			expErr:  generator.ErrorBadCharSet,
		},
		{
			desc:    "too long a characterset",
			charSet: edgeTooLongCharSet,
			expErr:  generator.ErrorBadCharSet,
		},
	}

	for _, tc := range testcases {
		g, err := generator.NewRandom(tc.charSet)
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
	g, err := generator.NewRandom(generator.AllChars)
	if err != nil {
		t.Fatalf("NewRandom: %s", err)
	}
	size := 36
	p, err := g.SecureRandomBytes(size)
	if err != nil {
		t.Fatalf("SecureRandomBytes: %s", err)
	}
	fmt.Printf("Password generated %s\n", string(p))
	if len(p) != size {
		t.Errorf("Expected password of length %d but got %d", size, len(p))
	}
}
