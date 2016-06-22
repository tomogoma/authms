package auth

import (
	"time"

	"bitbucket.org/tomogoma/auth-ms/auth/model/details/login"
	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/token"
	"bitbucket.org/tomogoma/auth-ms/auth/model/user"
)

const (
	numPrevLogins = 5
)

type Auth struct {
	usrM       *user.Model
	tokenM     *token.Model
	loginDetsM *login.Model
}

func New(dsnF helper.DSNFormatter, quitCh chan error) (*Auth, error) {

	db, err := helper.SQLDB(dsnF)
	if err != nil {
		return nil, err
	}

	usrM, err := user.NewModel(db)
	if err != nil {
		return nil, err
	}

	tokenM, err := token.NewModel(db)
	if err != nil {
		return nil, err
	}

	err = tokenM.RunGarbageCollector(quitCh)
	if err != nil {
		return nil, err
	}

	loginM, err := login.NewModel(db)
	if err != nil {
		return nil, err
	}

	err = helper.CreateTables(db, usrM, tokenM, loginM)
	if err != nil {
		return nil, err
	}

	return &Auth{usrM: usrM, tokenM: tokenM, loginDetsM: loginM}, nil
}

func (a *Auth) RegisterUser(usr user.User) (int, error) {

	return a.usrM.Save(usr)
}

func (a *Auth) Login(uName, devID, pass, rIP, srvID, ref string) (*user.User, error) {

	usr, err := a.usrM.Get(uName, pass)
	if err != nil {
		return nil, err
	}

	token, err := token.New(uName, devID, token.ShortExpType)
	if err != nil {
		return nil, err
	}
	usr.SetToken(token)

	_, err = a.tokenM.Save(*token)
	if err != nil {
		return nil, err
	}

	prevLogins, err := a.loginDetsM.Get(usr.UserName(), 0, numPrevLogins)
	if err != nil {
		return nil, err
	}
	usr.SetPreviousLogins(prevLogins)

	loginDets, err := login.New(usr.ID(), time.Now(), rIP, srvID, ref)
	if err != nil {
		a.tokenM.Delete(token.Token())
		return nil, err
	}

	_, err = a.loginDetsM.Save(*loginDets)
	if err != nil {
		a.tokenM.Delete(token.Token())
		return nil, err
	}

	return usr, nil
}

func (a *Auth) AuthenticateToken(usrID int, token string) (*user.User, error) {

	token, err := a.tokenM.Get(usrID, token)
	if err != nil {
		return nil, err
	}

	usr, err := a.usrM.Get(token.UserID())
	if err != nil {
		return nil, err
	}
	usr.SetToken(token)

	prevLogins, err := a.loginDetsM.Get(usr.UserName(), 0, numPrevLogins)
	if err != nil {
		return nil, err
	}
	usr.SetPreviousLogins(prevLogins)

	return usr, nil
}
