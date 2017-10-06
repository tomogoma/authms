package testing

import (
	"github.com/tomogoma/authms/service"
)

type GuardMock struct {
	ExpAPIKValidErr error
	ExpNewAPIK      *service.APIKey
	ExpNewAPIKErr   error
}

func (g *GuardMock) APIKeyValid(key string) error {
	return g.ExpAPIKValidErr
}
func (g *GuardMock) NewAPIKey(userID string) (*service.APIKey, error) {
	return g.ExpNewAPIK, g.ExpNewAPIKErr
}
