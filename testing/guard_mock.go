package testing

import (
	"github.com/tomogoma/authms/api"
)

type GuardMock struct {
	ExpAPIKValidUsrID string
	ExpAPIKValidErr   error
	ExpNewAPIK        *api.Key
	ExpNewAPIKErr     error
}

func (g *GuardMock) APIKeyValid(key string) (string, error) {
	return g.ExpAPIKValidUsrID, g.ExpAPIKValidErr
}
func (g *GuardMock) NewAPIKey(userID string) (*api.Key, error) {
	return g.ExpNewAPIK, g.ExpNewAPIKErr
}
