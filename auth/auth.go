package auth

import (
	"fmt"
	"strings"
	"time"

	"database/sql"
	"errors"

	"github.com/tomogoma/authms/auth/dbhelper/history"
	"github.com/tomogoma/authms/auth/dbhelper/token"
	"github.com/tomogoma/authms/auth/dbhelper/user"
	"github.com/tomogoma/authms/auth/oauth"
	"github.com/tomogoma/authms/auth/oauth/response"
	"github.com/tomogoma/authms/auth/dbhelper"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/auth/hash"
)

const (
	numPrevLogins = 5
)

var hashF = hash.Hasher{}
var ErrorNilHistoryModel = errors.New("history model was nil")
var ErrorNilTokenGenerator = errors.New("token generator was nil")
var ErrorNilPasswordGenerator = errors.New("password generator was nil")
var ErrorNilOAuthHandler = errors.New("oauth handler was nil")
var ErrorOAuthTokenNotValid = errors.New("oauth token is invalid")

type HistModel interface {
	Save(ld history.History) (int, error)
	Get(userID, offset, count int, acMs ...int) ([]*history.History, error)
}

type OAuthHandler interface {
	ValidateToken(appName, token string) (response.OAuth, error)
}

type TokenGenerator interface {
	Generate(usrID int, devID string, expType token.ExpiryType) (token.Token, error)
	Validate(tokenStr string) (token.Token, error)
}

type PasswordGenerator interface {
	SecureRandomString(length int) ([]byte, error)
}

type Auth struct {
	conf         Config
	usrM         *dbhelper.DBHelper
	tokenM       *token.Model
	histM        HistModel
	tokenG       TokenGenerator
	passwordG    PasswordGenerator
	oAuthHandler OAuthHandler
}

func AuthError(err error) bool {

	if strings.HasPrefix(err.Error(), dbhelper.ErrorPasswordMismatch.Error()) ||
		strings.HasPrefix(err.Error(), ErrorOAuthTokenNotValid.Error()) ||
		strings.HasPrefix(err.Error(), token.ErrorExpiredToken.Error()) ||
		strings.HasPrefix(err.Error(), token.ErrorInvalidToken.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmptyUserName.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmptyPhone.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmptyEmail.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorInvalidOAuth.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmptyPassword.Error()) ||
		strings.HasPrefix(err.Error(), token.ErrorEmptyDevID.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorUserExists.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmailExists.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorPhoneExists.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorAppIDExists.Error()) ||
		strings.HasPrefix(err.Error(), oauth.ErrorUnsupportedApp.Error()) {
		return true
	}

	return false
}

func New(db *sql.DB, histM HistModel, tg TokenGenerator, pg PasswordGenerator,
conf Config, lg token.Logger, oa OAuthHandler, quitCh chan error) (*Auth, error) {
	if histM == nil {
		return nil, ErrorNilHistoryModel
	}
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	if tg == nil {
		return nil, ErrorNilTokenGenerator
	}
	if pg == nil {
		return nil, ErrorNilPasswordGenerator
	}
	usrM, err := db.New(db)
	if err != nil {
		return nil, err
	}
	tokenM, err := token.NewModel(db)
	if err != nil {
		return nil, err
	}
	if oa == nil {
		return nil, ErrorNilOAuthHandler
	}
	err = tokenM.RunGarbageCollector(quitCh, lg)
	if err != nil {
		return nil, err
	}
	return &Auth{usrM: usrM, tokenM: tokenM, histM: histM, tokenG: tg,
		passwordG: pg, oAuthHandler: oa}, nil
}

func (a *Auth) RegisterUserName(userName, pass string) (dbhelper.User, error) {
	u, err := dbhelper.NewByUserName(userName, pass, hashF)
	if err != nil {
		return nil, err
	}
	return a.usrM.Save(*u)
}

func (a *Auth) RegisterEmail(email string, pass string) (dbhelper.User, error) {
	u, err := dbhelper.NewByEmail(email, pass, hashF)
	if err != nil {
		return nil, err
	}
	return a.usrM.Save(*u)
}

func (a *Auth) RegisterPhone(phone string, pass string) (dbhelper.User, error) {
	u, err := dbhelper.NewByPhone(phone, pass, hashF)
	if err != nil {
		return nil, err
	}
	return a.usrM.Save(*u)
}

func (a *Auth) RegisterOAuth(user *authms.User) error {
	if err := a.validateOAuth(user.OAuth); err != nil {
		return nil, err
	}
	return a.usrM.Save(user)
}

func (a *Auth) LoginUserName(uName, pass, devID, rIP, srvID, ref string) (dbhelper.User, error) {
	usr, err := a.usrM.GetByUserName(uName, pass, valHashF)
	if err != nil {
		if usr.ID() < 1 {
			return nil, err
		}
		return nil, a.saveHistory(usr.ID(), rIP, srvID, ref, history.LoginAccess, err)
	}
	tkn, err := a.tokenG.Generate(usr.ID(), devID, token.ShortExpType)
	if err != nil {
		return nil, err
	}
	usr.Token(tkn.Token())
	modelTkn, err := token.NewFrom(tkn)
	if err != nil {
		return nil, err
	}

	_, err = a.tokenM.Save(*modelTkn)
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

func (a *Auth) LoginOAuth(app dbhelper.App, devID, rIP, srvID, ref string) (dbhelper.User, error) {
	if err := a.validateOAuth(app); err != nil {
		return nil, err
	}
	usr, err := a.usrM.GetByAppUserID(app.Name(), app.UserID())
	if err != nil {
		if usr.ID() < 1 {
			return nil, err
		}
		return nil, a.saveHistory(usr.ID(), rIP, srvID, ref, history.LoginAccess, err)
	}
	tkn, err := a.tokenG.Generate(usr.ID(), devID, token.ShortExpType)
	if err != nil {
		return nil, err
	}
	usr.Token(tkn.Token())
	modelTkn, err := token.NewFrom(tkn)
	if err != nil {
		return nil, err
	}
	_, err = a.tokenM.Save(*modelTkn)
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

func (a *Auth) AuthenticateToken(tknStr, rIP, srvID, ref string) (dbhelper.User, error) {

	claimsTkn, err := a.tokenG.Validate(tknStr)
	if err != nil {
		return nil, err
	}

	tkn, err := a.tokenM.Get(claimsTkn.UserID(), claimsTkn.DevID(), tknStr)
	if err != nil {
		if tkn.UserID() < 1 {
			return nil, err
		}
		return nil, a.saveHistory(tkn.UserID(), rIP, srvID, ref, history.TokenValidationAccess, err)
	}

	usr, err := dbhelper.NewByToken(tkn.UserID(), tkn.Token())
	if err != nil {
		return nil, err
	}

	prevLogins, err := a.histM.Get(usr.ID(), 0, numPrevLogins, history.LoginAccess)
	if err != nil {
		if tkn.UserID() < 1 {
			return nil, err
		}
		return nil, a.saveHistory(tkn.UserID(), rIP, srvID, ref, history.TokenValidationAccess, err)
	}
	usr.SetPreviousLogins(prevLogins...)

	err = a.saveHistory(usr.ID(), rIP, srvID, ref, history.TokenValidationAccess, nil)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (a *Auth) validateOAuth(claimOA *authms.OAuth) error {
	oa, err := a.oAuthHandler.ValidateToken(claimOA.AppName, claimOA.AppToken)
	if err != nil {
		return err
	}
	if !oa.IsValid() || oa.UserID() != claimOA.AppUserID() {
		return ErrorOAuthTokenNotValid
	}
	return nil
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
