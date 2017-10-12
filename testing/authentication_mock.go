package testing

import (
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

type AuthenticationMock struct {
	errors.NotImplErrCheck
	errors.AuthErrCheck
	errors.ClErrCheck

	ExpRegSelfUser *model.User
	ExpRegSelfErr  error

	ExpRegSelfBLPUser *model.User
	ExpRegSelfBLPErr  error

	ExpRegOtherUser *model.User
	ExpRegOtherErr  error

	ExpUpdIDerUser *model.User
	ExpUpdIDerErr  error

	ExpLoginUser *model.User
	ExpLoginErr  error

	ExpSetPassVerLogin *model.VerifLogin
	ExpSetPassErr      error

	ExpVerDBTVerLogin *model.VerifLogin
	ExpVerDBTErr      error

	ExpSndVerCodeDBTStts *model.DBTStatus
	ExpSndVerCodeErr     error

	ExpSndPassRstDBTStts *model.DBTStatus
	ExpSndPassRstErr     error

	ExpVerExtDBT    string
	ExpVerExtDBTErr error

	ExpUpdPassErr error
}

func (a *AuthenticationMock) RegisterSelf(loginType, userType, id string, secret []byte) (*model.User, error) {
	return a.ExpRegSelfUser, a.ExpRegSelfErr
}

func (a *AuthenticationMock) RegisterSelfByLockedPhone(userType, devID, number string, secret []byte) (*model.User, error) {
	return a.ExpRegSelfBLPUser, a.ExpRegSelfBLPErr
}

func (a *AuthenticationMock) RegisterOther(JWT, newLoginType, userType, id, groupID string) (*model.User, error) {
	return a.ExpRegOtherUser, a.ExpRegOtherErr
}

func (a *AuthenticationMock) UpdateIdentifier(JWT, loginType, newId string) (*model.User, error) {
	return a.ExpUpdIDerUser, a.ExpUpdIDerErr
}

func (a *AuthenticationMock) UpdatePassword(JWT string, old, newPass []byte) error {
	return a.ExpUpdPassErr
}

func (a *AuthenticationMock) SetPassword(loginType, userID string, dbt, pass []byte) (*model.VerifLogin, error) {
	return a.ExpSetPassVerLogin, a.ExpSetPassErr
}

func (a *AuthenticationMock) SendVerCode(JWT, loginType, toAddr string) (*model.DBTStatus, error) {
	return a.ExpSndVerCodeDBTStts, a.ExpSndVerCodeErr
}

func (a *AuthenticationMock) SendPassResetCode(loginType, toAddr string) (*model.DBTStatus, error) {
	return a.ExpSndPassRstDBTStts, a.ExpSndPassRstErr
}

func (a *AuthenticationMock) VerifyAndExtendDBT(lt, usrID string, dbt []byte) (string, error) {
	return a.ExpVerExtDBT, a.ExpVerExtDBTErr
}

func (a *AuthenticationMock) VerifyDBT(loginType, userID string, dbt []byte) (*model.VerifLogin, error) {
	return a.ExpVerDBTVerLogin, a.ExpVerDBTErr
}

func (a *AuthenticationMock) Login(loginType, identifier string, password []byte) (*model.User, error) {
	return a.ExpLoginUser, a.ExpLoginErr
}
