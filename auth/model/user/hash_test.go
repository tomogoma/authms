package user_test

import (
	"testing"

	"github.com/tomogoma/authms/auth/model/user"
)

func TestCompareHash(t *testing.T) {

	type testcase struct {
		pass        string
		comparePass string
		expEqual    bool
		desc        string
	}

	tcases := []testcase{
		{
			desc:        "Passwords match",
			pass:        "pass",
			comparePass: "pass",
			expEqual:    true,
		},
		{
			desc:        "Password mismatch",
			pass:        "pass",
			comparePass: "passa",
			expEqual:    false,
		},
	}

	for _, c := range tcases {

		passHB, err := user.Hash(c.pass)
		if err != nil {
			t.Errorf("Test %s: user.Hash(): %s", c.desc, err)
			continue
		}

		equal := user.CompareHash(c.comparePass, passHB)
		if equal != c.expEqual {
			t.Errorf("Test %s: expect compare hash to be %t but got %t", c.desc, c.expEqual, equal)
		}

	}

}
