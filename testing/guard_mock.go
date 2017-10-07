package testing

import (
	"github.com/tomogoma/authms/service"
)

type GuardMock struct {
	ExpAPIKValidUsrID string
	ExpAPIKValidErr   error
	ExpNewAPIK        *service.APIKey
	ExpNewAPIKErr     error
}

func (g *GuardMock) APIKeyValid(key string) (string, error) {
	return g.ExpAPIKValidUsrID, g.ExpAPIKValidErr
}
func (g *GuardMock) NewAPIKey(userID string) (*service.APIKey, error) {
	return g.ExpNewAPIK, g.ExpNewAPIKErr
}
