package testing

import "github.com/tomogoma/authms/model"

type GuardMock struct {
	ExpAPIKValidErr error
	ExpNewAPIK      *model.APIKey
	ExpNewAPIKErr   error
}

func (g *GuardMock) APIKeyValid(userID, key string) error {
	return g.ExpAPIKValidErr
}
func (g *GuardMock) NewAPIKey(userID string) (*model.APIKey, error) {
	return g.ExpNewAPIK, g.ExpNewAPIKErr
}
