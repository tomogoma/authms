package auth

import (
	"fmt"
	"strings"
	"time"

	"database/sql"
	"errors"

	"github.com/tomogoma/authms/auth/model/history"
	"github.com/tomogoma/authms/auth/model/token"
	"github.com/tomogoma/authms/auth/model/user"
)

const (
	numPrevLogins = 5
)

var hashF = user.Hash
var valHashF = user.CompareHash
var ErrorNilHistoryModel = errors.New("history model was nil")

type HistModel interface {
	Save(ld history.History) (int, error)
	Get(userID, offset, count int, acMs ...int) ([]*history.History, error)
}

type User interface {
	UserName() string
	EmailAddress() string
	PhoneNumber() string
	App() user.App
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
		strings.HasPrefix(err.Error(), token.ErrorInvalidToken.Error()) ||
		strings.HasPrefix(err.Error(), user.ErrorEmptyIdentifier.Error()) ||
		strings.HasPrefix(err.Error(), user.ErrorEmptyPassword.Error()) ||
		strings.HasPrefix(err.Error(), token.ErrorEmptyDevID.Error()) ||
		strings.HasPrefix(err.Error(), user.ErrorUserExists.Error()) ||
		strings.HasPrefix(err.Error(), user.ErrorEmailExists.Error()) ||
		strings.HasPrefix(err.Error(), user.ErrorPhoneExists.Error()) ||
		strings.HasPrefix(err.Error(), user.ErrorAppIDExists.Error()) {
		return true
	}

	return false
}

func New(db *sql.DB, histM HistModel, conf Config, lg token.Logger, quitCh chan error) (*Auth, error) {

	if histM == nil {
		return nil, ErrorNilHistoryModel
	}

	if err := conf.Validate(); err != nil {
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

	err = tokenM.RunGarbageCollector(quitCh, lg)
	if err != nil {
		return nil, err
	}

	return &Auth{usrM: usrM, tokenM: tokenM, histM: histM}, nil
}

func (a *Auth) RegisterUser(usr User, pass string, rIP, srvID, ref string) (user.User, error) {

	u, err := user.New(usr.UserName(), usr.PhoneNumber(), usr.EmailAddress(), pass,
		usr.App(), hashF)
	fmt.Printf("created user: %+v", err)
	if err != nil {
		return nil, a.saveHistory(-1, rIP, srvID, ref, history.RegistrationAccess, err)
	}

	savedU, err := a.usrM.Save(*u)
	if err != nil {
		return nil, err
	}

	return savedU, a.saveHistory(savedU.ID(), rIP, srvID, ref, history.RegistrationAccess, nil)
}

func (a *Auth) LoginUserName(uName, pass, devID, rIP, srvID, ref string) (user.User, error) {

	usr, err := a.usrM.GetByUserName(uName, pass, valHashF)
	if err != nil {
		uid := -1
		if usr != nil {
			uid = usr.ID()
		}
		return nil, a.saveHistory(uid, rIP, srvID, ref, history.LoginAccess, err)
	}

	tkn, err := token.New(usr.ID(), devID, token.ShortExpType)
	if err != nil {
		return nil, err
	}
	usr.SetToken(tkn)

	_, err = a.tokenM.Save(*tkn)
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
		a.tokenM.Delete(tkn.Token())
		return nil, err
	}

	return usr, nil
}

func (a *Auth) AuthenticateToken(usrID int, devID, tknStr, rIP, srvID, ref string) (user.User, error) {

	tkn, err := a.tokenM.Get(usrID, devID, tknStr)
	if err != nil {
		if usrID == 0 {
			usrID = -1
		}
		return nil, a.saveHistory(usrID, rIP, srvID, ref, history.TokenValidationAccess, err)
	}

	usr, err := a.usrM.Get(tkn.UserID())
	if err != nil {
		return nil, err
	}
	usr.SetToken(tkn)

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
