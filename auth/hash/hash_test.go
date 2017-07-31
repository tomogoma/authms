package hash_test

import (
	"github.com/tomogoma/authms/auth/hash"
	"testing"
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

	h := hash.Hasher{}
	for _, c := range tcases {

		passHB, err := h.Hash(c.pass)
		if err != nil {
			t.Errorf("Test %s: user.Hash(): %s", c.desc, err)
			continue
		}

		equal := h.CompareHash(c.comparePass, passHB)
		if equal != c.expEqual {
			t.Errorf("Test %s: expect compare hash to be %t but got %t", c.desc, c.expEqual, equal)
		}

	}

}
