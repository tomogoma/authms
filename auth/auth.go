package auth

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/alkira/contactsms/kazoo/errors"
	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/history"
	"bitbucket.org/tomogoma/auth-ms/auth/model/token"
	"bitbucket.org/tomogoma/auth-ms/auth/model/user"
)

const (
	numPrevLogins = 5
)

var hashF = user.Hash
var valHashF = user.CompareHash
var ErrorNilHistoryModel = errors.New("history model was nil")

type HistModel interface {
	helper.Model
	Save(ld history.History) (int, error)
	Get(userID, offset, count int, acMs ...int) ([]*history.History, error)
}

type User interface {
	UserName() string
	FirstName() string
	MiddleName() string
	LastName() string
}

type Auth struct {
	conf   Config
	usrM   *user.Model
	tokenM *token.Model
	histM  HistModel
}

func AuthError(err error) bool {

	if strings.HasPrefix(err.Error(), user.ErrorPasswordMismatch.Error()) ||
		strings.HasPrefix(err.Error(), token.ErrorExpiredToken.Error()) ||
		strings.HasPrefix(err.Error(), user.ErrorUserExists.Error()) ||
		strings.HasPrefix(err.Error(), token.ErrorInvalidToken.Error()) ||
		strings.HasPrefix(err.Error(), user.ErrorEmptyUserName.Error()) ||
		strings.HasPrefix(err.Error(), user.ErrorEmptyPassword.Error()) {
		return true
	}

	return false
}

func New(dsnF helper.DSNFormatter, histM HistModel, conf Config, lg token.Logger, quitCh chan error) (*Auth, error) {

	if histM == nil {
		return nil, ErrorNilHistoryModel
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

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

	err = helper.CreateTables(db, usrM, tokenM, histM)
	if err != nil {
		return nil, err
	}

	err = tokenM.RunGarbageCollector(quitCh, lg)
	if err != nil {
		return nil, err
	}

	return &Auth{usrM: usrM, tokenM: tokenM, histM: histM}, nil
}

func (a *Auth) RegisterUser(usr User, pass string, rIP, srvID, ref string) (user.User, error) {

	u, err := user.New(usr.UserName(), usr.FirstName(), usr.MiddleName(),
		usr.LastName(), pass, hashF)
	if err != nil {
		return nil, a.saveHistory(-1, rIP, srvID, ref, history.RegistrationAccess, err)
	}

	savedU, err := a.usrM.Save(*u)
	if err != nil {
		return nil, err
	}

	return savedU, a.saveHistory(savedU.ID(), rIP, srvID, ref, history.RegistrationAccess, nil)
}

func (a *Auth) Login(uName, pass, devID, rIP, srvID, ref string) (user.User, error) {

	usr, err := a.usrM.Get(uName, pass, valHashF)
	if err != nil {
		uid := -1
		if usr != nil {
			uid = usr.ID()
		}
		return nil, a.saveHistory(uid, rIP, srvID, ref, history.LoginAccess, err)
	}

	token, err := token.New(usr.ID(), devID, token.ShortExpType)
	if err != nil {
		return nil, err
	}
	usr.SetToken(token)

	_, err = a.tokenM.Save(*token)
	if err != nil {
		return nil, err
	}

	prevLogins, err := a.histM.Get(usr.ID(), 0, numPrevLogins, history.LoginAccess)
	if err != nil {
		return nil, err
	}
	usr.SetPreviousLogins(prevLogins...)

	err = a.saveHistory(usr.ID(), rIP, srvID, ref, history.LoginAccess, nil)
	if err != nil {
		a.tokenM.Delete(token.Token())
		return nil, err
	}

	return usr, nil
}

func (a *Auth) AuthenticateToken(usrID int, devID, tknStr, rIP, srvID, ref string) (user.User, error) {

	token, err := a.tokenM.Get(usrID, devID, tknStr)
	if err != nil {
		return nil, a.saveHistory(usrID, rIP, srvID, ref, history.TokenValidationAccess, err)
	}

	usr, err := a.usrM.GetByID(token.UserID())
	if err != nil {
		return nil, err
	}
	usr.SetToken(token)

	prevLogins, err := a.histM.Get(usr.ID(), 0, numPrevLogins, history.LoginAccess)
	if err != nil {
		return nil, err
	}
	usr.SetPreviousLogins(prevLogins...)

	err = a.saveHistory(usr.ID(), rIP, srvID, ref, history.TokenValidationAccess, nil)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (a *Auth) saveHistory(id int, rIP, srvID, ref string, accType int, err error) error {

	accSuccessful := true
	if err != nil {
		accSuccessful = false
	}

	h, hErr := history.New(id, accType, accSuccessful, time.Now(), rIP, srvID, ref)
	if hErr != nil {
		if err != nil {
			return fmt.Errorf("%s ...further error saving history: %s", err, hErr)
		}
		return hErr
	}

	_, hErr = a.histM.Save(*h)
	if hErr != nil {
		if err != nil {
			return fmt.Errorf("%s ...further error saving history: %s", err, hErr)
		}
		return hErr
	}

	return err
}
