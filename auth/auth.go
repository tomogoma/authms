package auth

import (
	"strings"

	"github.com/tomogoma/authms/auth/oauth"
	"github.com/tomogoma/authms/auth/oauth/response"
	"github.com/tomogoma/authms/auth/dbhelper"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/go-commons/errors"
)

type OAuthHandler interface {
	ValidateToken(appName, token string) (response.OAuth, error)
}

type TokenGenerator interface {
	Generate(usrID int, devID string, expType token.ExpiryType) (*token.Token, error)
	Validate(tokenStr string) (*token.Token, error)
	ValidateUser(tokenStr string, userID int) (*token.Token, error)
}

type PasswordGenerator interface {
	SecureRandomString(length int) ([]byte, error)
}

type Logger interface {
	Info(interface{}, ...interface{})
	Error(interface{}, ...interface{}) error
}

type DBHelper interface {
	SaveUser(*authms.User) error
	GetByUserName(uname, pass string) (*authms.User, error)
	GetByAppUserID(appName, appUserID string) (*authms.User, error)
	SaveToken(*token.Token) error
	GetHistory(userID int64, offset, count int, accessType ...string) ([]*authms.History, error)
	SaveHistory(*authms.History) error
	UpdatePhone(userID int64, newPhone *authms.Value) error
	UpdateAppUserID(userID int64, new *authms.OAuth) error
}

type Auth struct {
	dbHelper     DBHelper
	tokenG       TokenGenerator
	logger       Logger
	oAuthHandler OAuthHandler
}

const (
	numPrevLogins = 5
)

var ErrorNilTokenGenerator = errors.New("token generator was nil")
var ErrorNilLogger = errors.New("Logger was nil")
var ErrorNilDBHelper = errors.New("DBHelper was nil")
var ErrorNilOAuthHandler = errors.New("oauth handler was nil")
var ErrorOAuthTokenNotValid = errors.New("oauth token is invalid")

func AuthError(err error) bool {
	if strings.HasPrefix(err.Error(), dbhelper.ErrorPasswordMismatch.Error()) ||
		strings.HasPrefix(err.Error(), ErrorOAuthTokenNotValid.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmptyUserName.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmptyPhone.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmptyEmail.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorInvalidOAuth.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmptyPassword.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorUserExists.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorEmailExists.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorPhoneExists.Error()) ||
		strings.HasPrefix(err.Error(), dbhelper.ErrorAppIDExists.Error()) ||
		strings.HasPrefix(err.Error(), oauth.ErrorUnsupportedApp.Error()) {
		return true
	}
	return false
}

func New(tg TokenGenerator, lg Logger, db DBHelper,
oa OAuthHandler) (*Auth, error) {
	if tg == nil {
		return nil, ErrorNilTokenGenerator
	}
	if lg == nil {
		return nil, ErrorNilLogger
	}
	if db == nil {
		return nil, ErrorNilDBHelper
	}
	if oa == nil {
		return nil, ErrorNilOAuthHandler
	}
	return &Auth{dbHelper: db, tokenG: tg, oAuthHandler: oa, logger: lg}, nil
}

func (a *Auth) Register(user *authms.User, devID, rIP string) error {
	if user == nil {
		return errors.NewClient("user was empty")
	}
	if devID == "" {
		return errors.NewClient("Dev ID was empty")
	}
	for appName, oa := range user.OAuths {
		if appName != oa.AppName {
			return errors.NewClient("an OAuth's key does not match its app name")
		}
		if err := a.validateOAuth(oa); err != nil {
			return err
		}
		user.OAuths[appName].Verified = false
	}
	if user.Phone != nil {
		user.Phone.Verified = false
	}
	if user.Email != nil {
		user.Email.Verified = false
	}
	err := a.dbHelper.SaveUser(user)
	if err != nil {
		return errors.Newf("error persisting user: %v", err)
	}
	go a.saveHistory(user, devID, dbhelper.AccessRegistration, rIP, nil)
	return nil
}

func (a *Auth) UpdatePhone(user *authms.User, token, devID, rIP string) error {
	if user == nil {
		return errors.NewClient("user was empty")
	}
	_, err := a.tokenG.ValidateUser(token, int(user.ID))
	defer func() {
		go a.saveHistory(user, devID, dbhelper.AccessUpdate, rIP, err)
	}()
	if err != nil {
		return err
	}
	if !dbhelper.HasValue(user.Phone) {
		return errors.NewClient("phone was invalid")
	}
	if devID == "" {
		return errors.NewClient("device ID was empty")
	}
	user.Phone.Verified = false
	err = a.dbHelper.UpdatePhone(user.ID, user.Phone)
	if err != nil {
		return errors.Newf("error persisting phone update: %v", err)
	}
	return nil
}

func (a *Auth) UpdateOAuth(user *authms.User, appName, token, devID, rIP string) error {
	if user == nil {
		return errors.NewClient("user was empty")
	}
	_, err := a.tokenG.ValidateUser(token, int(user.ID))
	defer func() {
		go a.saveHistory(user, devID, dbhelper.AccessUpdate, rIP, err)
	}()
	if err != nil {
		return err
	}
	if user.OAuths == nil || user.OAuths[appName] == nil {
		return errors.NewClient("OAuth was not provided")
	}
	oa := user.OAuths[appName]
	if err := a.validateOAuth(oa); err != nil {
		return err
	}
	if devID == "" {
		return errors.NewClient("device ID was empty")
	}
	oa.Verified = true
	err = a.dbHelper.UpdateAppUserID(user.ID, oa)
	if err != nil {
		return errors.Newf("error persisting OAuth changes: %v", err)
	}
	return nil
}

func (a *Auth) UpdatePassword(userID int64, oldPass, newPass, devID, rIP string) error {
	return errors.New("not yet implemented")
}

func (a *Auth) LoginUserName(uName, pass, devID, rIP string) (*authms.User, error) {
	if devID == "" {
		return nil, errors.NewClient("Dev ID was empty")
	}
	usr, err := a.dbHelper.GetByUserName(uName, pass)
	if err = a.processLoginResults(usr, devID, rIP, err); err != nil {
		return nil, err
	}
	return usr, nil
}

func (a *Auth) LoginOAuth(app *authms.OAuth, devID, rIP string) (*authms.User, error) {
	if devID == "" {
		return nil, errors.NewClient("Dev ID was empty")
	}
	if err := a.validateOAuth(app); err != nil {
		return nil, err
	}
	usr, err := a.dbHelper.GetByAppUserID(app.AppName, app.AppUserID)
	if err = a.processLoginResults(usr, devID, rIP, err); err != nil {
		return nil, err
	}
	return usr, nil
}

func (a *Auth) processLoginResults(usr *authms.User, devID, rIP string, loginErr error) error {
	defer func() {
		go a.saveHistory(usr, devID, dbhelper.AccessLogin, rIP, loginErr)
	}()
	if loginErr != nil {
		return loginErr
	}
	tkn, loginErr := a.tokenG.Generate(int(usr.ID), devID, token.ShortExpType)
	if loginErr != nil {
		loginErr = errors.Newf("error generating token: %v", loginErr)
		return loginErr
	}
	loginErr = a.dbHelper.SaveToken(tkn)
	if loginErr != nil {
		loginErr = errors.Newf("error persisting login token: %v", loginErr)
		return loginErr
	}
	usr.Token = tkn.Token()
	prevLogins, loginErr := a.dbHelper.GetHistory(usr.ID, 0, numPrevLogins,
		dbhelper.AccessLogin)
	if loginErr != nil {
		loginErr = errors.Newf("error fetching login history: %v", loginErr)
		return loginErr
	}
	usr.LoginHistory = prevLogins
	return nil
}

func (a *Auth) validateOAuth(claimOA *authms.OAuth) error {
	if claimOA == nil {
		return nil
	}
	oa, err := a.oAuthHandler.ValidateToken(claimOA.AppName, claimOA.AppToken)
	if err != nil {
		return errors.Newf("error validating OAuth token: %v", err)
	}
	if !oa.IsValid() || oa.UserID() != claimOA.AppUserID {
		return ErrorOAuthTokenNotValid
	}
	return nil
}

func (a *Auth) saveHistory(user *authms.User, devID, accType, rIP string, err error) {
	if user == nil || user.ID < 1 {
		return
	}
	accSuccessful := true
	if err != nil {
		accSuccessful = false
	}
	h := &authms.History{UserID: user.ID, AccessType: accType,
		SuccessStatus: accSuccessful, IpAddress: rIP, DevID: devID}
	err = a.dbHelper.SaveHistory(h)
	if err != nil {
		a.logger.Error("unable to save auth history entry (' %+v '): %s", h, err)
	}
}
