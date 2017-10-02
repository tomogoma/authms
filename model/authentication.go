package model

import (
	"bytes"
	"database/sql"
	"html/template"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/go-commons/errors"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

type AuthStore interface {
	IsNotFoundError(error) bool
	ExecuteTx(fn func(*sql.Tx) error) error

	GroupByName(string) (*Group, error)
	Group(string) (*Group, error)
	InsertGroup(name string, acl int) (*Group, error)
	AddUserToGroupAtomic(tx *sql.Tx, userID, groupID string) error

	UserTypeByName(string) (*UserType, error)
	InsertUserType(name string) (*UserType, error)
	InsertUserAtomic(tx *sql.Tx, typeID string, password []byte) (*User, error)

	UpdateUserPhone(userID, phone string, verified bool) (*VerifLogin, error)
	UpdateUserEmail(userID, email string, verified bool) (*VerifLogin, error)

	UpdatePassword(userID string, password []byte) error
	UpdatePasswordAtomic(tx *sql.Tx, userID string, password []byte) error
	UpdateUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*VerifLogin, error)
	UpdateUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*VerifLogin, error)

	InsertUserPhone(userID, phone string, verified bool) (*VerifLogin, error)
	InsertUserEmail(userID, email string, verified bool) (*VerifLogin, error)
	InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*VerifLogin, error)
	InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*VerifLogin, error)
	InsertUserNameAtomic(tx *sql.Tx, userID, username string) (*Username, error)
	InsertUserFacebookIDAtomic(tx *sql.Tx, userID, fbID string, verified bool) (*Facebook, error)

	InsertPhoneToken(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	InsertEmailToken(userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	InsertPhoneTokenAtomic(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	InsertEmailTokenAtomic(tx *sql.Tx, userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	PhoneTokens(userID string, offset, count int64) ([]*DBToken, error)
	EmailTokens(userID string, offset, count int64) ([]*DBToken, error)

	User(id string) (*User, []byte, error)
	UserByPhone(phone string) (*User, []byte, error)
	UserByEmail(email string) (*User, []byte, error)
	UserByUserName(username string) (*User, []byte, error)
	UserByFacebook(facebookID string) (*User, error)
}

type Guard interface {
	APIKeyValid(userID, key string) error
	NewAPIKey(userID string) (*APIKey, error)
}

type SecureRandomByteser interface {
	SecureRandomBytes(length int) ([]byte, error)
}

type FacebookCl interface {
	IsAuthError(error) bool
	ValidateToken(string) (FacebookResponse, error)
}

type SMSer interface {
	SMS(toPhone, message string) error
}

type JWTEr interface {
	Generate(claims jwt.Claims) (string, error)
	Validate(jwt string, claims jwt.Claims) (*jwt.Token, error)
}

type Mailer interface {
	SetConfig(conf SMTPConfig, notifEmail SendMail) error
	SendEmail(email SendMail) error
}

type APIKey struct {
	ID         string
	UserID     string
	APIKey     string
	CreateDate time.Time
	UpdateDate time.Time
}

type Authentication struct {
	appName   string
	webAppURL string

	allowSelfReg bool
	guard        Guard
	db           AuthStore
	fb           FacebookCl
	passGen      SecureRandomByteser
	numGen       SecureRandomByteser
	urlTokenGen  SecureRandomByteser
	smser        SMSer
	jwter        JWTEr
	mailer       Mailer

	invitationSubject   string
	verificationSubject string
	resetPassSubject    string
	loginTpActionTplts  map[string]map[string]*template.Template
}

const (
	GroupSuper  = "super"
	GroupAdmin  = "admin"
	GroupStaff  = "staff"
	GroupPublic = "public"

	AccessLevelSuper  = 0
	AccessLevelAdmin  = 3
	AccessLevelStaff  = 7
	AccessLevelPublic = 10

	UserTypeIndividual = "individual"
	UserTypeCompany    = "company"

	rePhoneChars = `[0-9]+`

	minPassLen = 8
	genPassLen = 32

	inviteValidity = 24 * time.Hour * 30
	resetValidity  = 2 * time.Hour
	verifyValidity = 5 * time.Minute

	ActionInvite    = "invite"
	ActionVerify    = "verify"
	ActionResetPass = "reset/password"

	loginTypeUsername = "usernames"
	loginTypeEmail    = "phones"
	loginTypePhone    = "emails"
	loginTypeFacebook = "facebook"
)

var (
	rePhone        = regexp.MustCompile(rePhoneChars)
	validUserTypes = []string{UserTypeIndividual, UserTypeCompany}

	ActionNotSupportedErrorF    = "action not supported for request: %s"
	LoginTypeNotSupportedErrorF = "login type not supported for request: %s"
)

func NewAuthentication() (*Authentication, error) {
	a := Authentication{}
	a.loginTpActionTplts = map[string]map[string]*template.Template{
		loginTypePhone: make(map[string]*template.Template),
		loginTypeEmail: make(map[string]*template.Template),
	}
	var err error
	if _, ok := a.loginTpActionTplts[loginTypePhone][ActionInvite]; !ok {
		a.loginTpActionTplts[loginTypePhone][ActionInvite], err =
			template.ParseFiles(config.DefaultPhoneInviteTpl)
		if err != nil {
			return nil, errors.Newf("read SMS invitation template: %v", err)
		}
	}
	if _, ok := a.loginTpActionTplts[loginTypePhone][ActionVerify]; !ok {
		a.loginTpActionTplts[loginTypePhone][ActionVerify], err =
			template.ParseFiles(config.DefaultPhoneVerifyTpl)
		if err != nil {
			return nil, errors.Newf("read SMS verification template: %v", err)
		}
	}
	if _, ok := a.loginTpActionTplts[loginTypePhone][ActionResetPass]; !ok {
		a.loginTpActionTplts[loginTypePhone][ActionResetPass], err =
			template.ParseFiles(config.DefaultPhoneResetPassTpl)
		if err != nil {
			return nil, errors.Newf("read SMS reset password template: %v", err)
		}
	}
	if _, ok := a.loginTpActionTplts[loginTypeEmail][ActionInvite]; !ok {
		a.loginTpActionTplts[loginTypeEmail][ActionInvite], err =
			template.ParseFiles(config.DefaultEmailInviteTpl)
		if err != nil {
			return nil, errors.Newf("read email invitation template: %v", err)
		}
	}
	if _, ok := a.loginTpActionTplts[loginTypeEmail][ActionVerify]; !ok {
		a.loginTpActionTplts[loginTypeEmail][ActionVerify], err =
			template.ParseFiles(config.DefaultEmailVerifyTpl)
		if err != nil {
			return nil, errors.Newf("read email verification template: %v", err)
		}
	}
	if _, ok := a.loginTpActionTplts[loginTypeEmail][ActionResetPass]; !ok {
		a.loginTpActionTplts[loginTypeEmail][ActionResetPass], err =
			template.ParseFiles(config.DefaultEmailResetPassTpl)
		if err != nil {
			return nil, errors.Newf("read email reset password template: %v", err)
		}
	}
	return nil, errors.NewNotImplemented()
}

func (a *Authentication) RegisterPublicByUsername(clientID, apiKey, userType, username, password string) (*User, error) {
	return a.registerUser(clientID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		if username == "" {
			return errors.NewClient("userName cannot be empty")
		}
		uname, err := a.db.InsertUserNameAtomic(tx, usr.ID, username)
		if err != nil {
			return errors.Newf("insert username: %v", err)
		}
		usr.UserName = *uname
		return nil
	})
}

func (a *Authentication) RegisterPublicByPhone(clientID, apiKey, userType, number, password string) (*User, error) {
	return a.registerUser(clientID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		if err := a.insertPhoneAtomic(tx, usr, number); err != nil {
			return err
		}
		if a.smser == nil {
			return nil
		}
		_, err := a.genAndSendTokens(tx, ActionVerify, loginTypePhone, number, usr.ID)
		return err
	})
}

func (a *Authentication) RegisterPublicByEmail(clientID, apiKey, userType, address, password string) (*User, error) {
	return a.registerUser(clientID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		if err := a.insertEmailAtomic(tx, usr, address); err != nil {
			return err
		}
		if a.mailer == nil {
			return nil
		}
		_, err := a.genAndSendTokens(tx, ActionVerify, loginTypeEmail, address, usr.ID)
		return err
	})
}

func (a *Authentication) RegisterPublicByFacebook(clientID, apiKey, userType, fbToken string) (*User, error) {
	if a.fb == nil {
		return nil, errors.NewNotImplementedf("facebook registration not available")
	}
	passwordB, err := a.passGen.SecureRandomBytes(genPassLen)
	if err != nil {
		return nil, errors.Newf("generate password: %v", err)
	}
	return a.registerUser(clientID, apiKey, userType, string(passwordB), func(tx *sql.Tx, usr *User) error {
		if fbToken == "" {
			return errors.NewClient("facebook token cannot be empty")
		}
		fbID, err := a.validateFbToken(fbToken)
		if err != nil {
			return err
		}
		fb, err := a.db.InsertUserFacebookIDAtomic(tx, usr.ID, fbID, true)
		if err != nil {
			return errors.Newf("insert facebook: %v", err)
		}
		usr.Facebook = *fb
		return nil
	})
}

func (a *Authentication) CreateByPhone(clientID, apiKey, jwt, userType, number, groupID string) (*User, error) {
	if a.smser == nil {
		return nil, errors.NewNotImplementedf("SMS notification (to created user) not available")
	}
	return a.createUser(clientID, apiKey, jwt, userType, groupID, func(tx *sql.Tx, usr *User) error {
		if err := a.insertPhoneAtomic(tx, usr, number); err != nil {
			return err
		}
		_, err := a.genAndSendTokens(tx, ActionInvite, loginTypePhone, number, usr.ID)
		return err
	})
}

func (a *Authentication) CreateByEmail(clientID, apiKey, jwt, userType, address, groupID string) (*User, error) {
	if a.mailer == nil {
		return nil, errors.NewNotImplementedf("email notification (to created user) not available")
	}
	return a.createUser(clientID, apiKey, jwt, userType, groupID, func(tx *sql.Tx, usr *User) error {
		if err := a.insertEmailAtomic(tx, usr, address); err != nil {
			return err
		}
		_, err := a.genAndSendTokens(tx, ActionInvite, loginTypeEmail, address, usr.ID)
		return err
	})
}

func (a *Authentication) UpdatePassword(clientID, apiKey, jwt, old, new string) error {

	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return err
	}

	clm := new(Claim)
	if _, err := a.jwter.Validate(jwt, clm); err != nil {
		return err
	}
	_, oldPassH, err := a.db.User(clm.UsrID)
	if err != nil {
		return errors.Newf("get user: %v", err)
	}

	if err = passwordValid(oldPassH, []byte(old)); err != nil {
		return err
	}
	newPassH, err := hashIfValid(new)
	if err != nil {
		return err
	}
	if err = a.db.UpdatePassword(clm.UsrID, newPassH); err != nil {
		return errors.Newf("update password: %v", err)
	}

	return nil
}

func (a *Authentication) SetPassword(clientID, apiKey, loginType, userID, dbt, pass string) (*VerifLogin, error) {

	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}

	var tkn *DBToken
	var err error
	var updtVerifiedFunc func(*sql.Tx, string, string, bool) (*VerifLogin, error)

	switch loginType {
	case loginTypeEmail:
		updtVerifiedFunc = a.db.UpdateUserEmailAtomic
		tkn, err = a.dbTokenValid(userID, dbt, a.db.EmailTokens)
	case loginTypePhone:
		updtVerifiedFunc = a.db.UpdateUserPhoneAtomic
		tkn, err = a.dbTokenValid(userID, dbt, a.db.PhoneTokens)
	default:
		return nil, errors.NewClientf(LoginTypeNotSupportedErrorF, loginType)
	}

	if err != nil {
		return nil, err
	}

	passH, err := hashIfValid(pass)
	if err != nil {
		return nil, err
	}

	var addr *VerifLogin
	err = a.db.ExecuteTx(func(tx *sql.Tx) error {
		err = a.db.UpdatePasswordAtomic(tx, userID, passH)
		if err != nil {
			return errors.Newf("update password: %v", err)
		}
		addr, err = updtVerifiedFunc(tx, userID, tkn.Phone, true)
		if err != nil {
			return errors.Newf("update phone to verified: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (a *Authentication) SendVerCode(clientID, apiKey, loginType, jwt, toAddr string) (string, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return "", err
	}

	clms := new(Claim)
	if _, err := a.jwter.Validate(jwt, clms); err != nil {
		return "", err
	}

	var err error
	var usr *User
	var insFunc func(userID, phone string, verified bool) (*VerifLogin, error)

	switch loginType {
	case loginTypePhone:
		insFunc = a.db.InsertUserPhone
		usr, _, err = a.db.UserByPhone(toAddr)
	case loginTypeEmail:
		insFunc = a.db.InsertUserEmail
		usr, _, err = a.db.UserByEmail(toAddr)
	default:
		return "", errors.NewClientf(LoginTypeNotSupportedErrorF, loginType)
	}

	if err != nil {
		if !a.db.IsNotFoundError(err) {
			return "", errors.Newf("user by %s: %v", loginType, err)
		}

		usr, _, err = a.db.User(clms.UsrID)
		if err != nil {
			if a.db.IsNotFoundError(err) {
				return "", errors.Newf("valid claims with non-exist user found: %+v", clms)
			}
			return "", errors.Newf("get user: %v", err)
		}

		_, err = insFunc(usr.ID, toAddr, false)
		if err != nil {
			return "", errors.Newf("insert toAddr: %v", err)
		}

	} else if usr.ID != clms.UsrID {
		return "", errors.NewForbiddenf("%s does not belong to you: %v", loginType)
	}

	return a.genAndSendTokens(nil, ActionVerify, loginType, toAddr, usr.ID)
}

func (a *Authentication) SendPassResetCode(clientID, apiKey, loginType, toAddr string) (string, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return "", err
	}
	var err error
	var usr *User

	switch loginType {
	case loginTypePhone:
		usr, _, err = a.db.UserByPhone(toAddr)
	case loginTypeEmail:
		usr, _, err = a.db.UserByEmail(toAddr)
	default:
		return "", errors.NewClientf(LoginTypeNotSupportedErrorF, loginType)
	}

	if err != nil {
		if a.db.IsNotFoundError(err) {
			return "", errors.NewForbiddenf("%s does not exist", toAddr)
		}
		return "", errors.Newf("user by %s: %v", loginType, err)
	}

	return a.genAndSendTokens(nil, ActionResetPass, loginType, toAddr, usr.ID)
}

func (a *Authentication) Verify(clientID, apiKey, loginType, userID, dbt string) (*VerifLogin, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}

	var tokensFetchFunc func(string, int64, int64) ([]*DBToken, error)
	var updateLoginFunc func(string, string, bool) (*VerifLogin, error)

	switch loginType {
	case loginTypeEmail:
		tokensFetchFunc = a.db.EmailTokens
		updateLoginFunc = a.db.UpdateUserEmail
	case loginTypePhone:
		tokensFetchFunc = a.db.PhoneTokens
		updateLoginFunc = a.db.UpdateUserPhone
	default:
		return nil, errors.NewClientf(LoginTypeNotSupportedErrorF, loginType)
	}

	tkn, err := a.dbTokenValid(userID, dbt, tokensFetchFunc)
	if err != nil {
		return nil, err
	}
	phone, err := updateLoginFunc(userID, tkn.Phone, true)
	if err != nil {
		return nil, errors.Newf("update phone to verified: %v", err)
	}
	return phone, nil
}

func (a *Authentication) Login(clientID, apiKey, loginType, identifier, password string) (*User, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}

	var usr *User
	var passHB []byte
	var err error

	switch loginType {
	case loginTypePhone:
		usr, passHB, err = a.db.UserByPhone(identifier)
	case loginTypeEmail:
		usr, passHB, err = a.db.UserByEmail(identifier)
	case loginTypeUsername:
		usr, passHB, err = a.db.UserByUserName(identifier)
	case loginTypeFacebook:
		return a.loginFacebook(identifier)
	default:
		return nil, errors.NewClientf(LoginTypeNotSupportedErrorF, loginType)
	}
	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewUnauthorizedf("invalid username/password combination")
		}
		return nil, errors.Newf("get user by %s: %v", loginType, err)
	}

	if err := passwordValid(passHB, []byte(password)); err != nil {
		return nil, err
	}
	return usr, nil
}

func (a *Authentication) UpdatePhoneLogin(c context.Context, userID, phone string) (*User, error) {
	return nil, errors.NewNotImplemented()
}

func (a *Authentication) genAndSendTokens(tx *sql.Tx, action, loginType, toAddr, usrID string) (string, error) {

	tkn, err := a.genAndInsertToken(tx, action, loginType, usrID, toAddr)
	if err != nil {
		return "", err
	}

	URL, err := a.genURL(action, loginType, usrID, tkn)
	if err != nil {
		return "", err
	}

	code, err := a.genAndInsertCode(tx, action, loginType, usrID, toAddr)
	if err != nil {
		return "", err
	}

	var sendData interface{}
	var subj string

	switch action {
	case ActionInvite:
		sendData = InvitationTemplate{AppName: a.appName, URLToken: URL}
		subj = a.invitationSubject
	case ActionVerify:
		sendData = VerificationTemplate{AppName: a.appName, URLToken: URL,
			Code: string(code)}
		subj = a.verificationSubject
	case ActionResetPass:
		sendData = VerificationTemplate{AppName: a.appName, URLToken: URL,
			Code: string(code)}
		subj = a.resetPassSubject
	default:
		return "", errors.Newf(ActionNotSupportedErrorF, action)
	}

	var obfuscateFunc func(string) string
	tpl := a.loginTpActionTplts[loginType][action]

	switch loginType {
	case loginTypePhone:
		obfuscateFunc = obfuscatePhone
		err = a.sendSMS(toAddr, tpl, sendData)
	case loginTypeEmail:
		obfuscateFunc = obfuscateEmail
		err = a.sendEmail(toAddr, subj, tpl, sendData)
	default:
		return "", errors.Newf(LoginTypeNotSupportedErrorF, loginType)
	}
	if err != nil {
		return "", err
	}

	return obfuscateFunc(toAddr), nil
}

func (a *Authentication) registerUser(clientID, apiKey, userType,
	password string, f func(tx *sql.Tx, usr *User) error) (*User, error) {

	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	if !a.allowSelfReg {
		return nil, errors.NewForbidden("registration closed to the public")
	}

	if !inStrs(userType, validUserTypes) {
		return nil, errors.NewClientf("accountType must be one of %+v", validUserTypes)
	}

	grp, err := a.getOrCreateGroup(GroupPublic, AccessLevelPublic)
	if err != nil {
		return nil, err
	}
	ut, err := a.getOrCreateUserType(userType)
	if err != nil {
		return nil, err
	}
	passH, err := hashIfValid(password)
	if err != nil {
		return nil, err
	}

	usr := new(User)
	err = a.db.ExecuteTx(func(tx *sql.Tx) error {
		usr, err = a.db.InsertUserAtomic(tx, ut.ID, passH)
		if err != nil {
			return errors.Newf("insert user: %v", err)
		}
		if err := a.db.AddUserToGroupAtomic(tx, usr.ID, grp.ID); err != nil {
			return errors.Newf("add user to group: %v", err)
		}
		if err := f(tx, usr); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	usr.Groups = []Group{*grp}
	usr.Type = *ut

	return usr, nil
}

func (a *Authentication) createUser(clientID, apiKey, jwt, userType,
	groupID string, f func(tx *sql.Tx, usr *User) error) (*User, error) {

	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	adminGrp, err := a.getOrCreateGroup(GroupAdmin, AccessLevelAdmin)
	if err != nil {
		return nil, err
	}
	if _, err := a.validateJWTInGroup(jwt, *adminGrp); err != nil {
		return nil, err
	}

	if !inStrs(userType, validUserTypes) {
		return nil, errors.NewClientf("accountType must be one of %+v", validUserTypes)
	}
	usrGroup, err := a.db.Group(groupID)
	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewClient("groupID does not exist")
		}
		return nil, errors.Newf("get group by ID: %v", err)
	}
	ut, err := a.getOrCreateUserType(userType)
	if err != nil {
		return nil, err
	}

	_, passH, err := a.genPasswordWithHash()
	if err != nil {
		return nil, err
	}
	usr := new(User)
	err = a.db.ExecuteTx(func(tx *sql.Tx) error {
		usr, err = a.db.InsertUserAtomic(tx, ut.ID, passH)
		if err != nil {
			return errors.Newf("insert user: %v", err)
		}
		if err := a.db.AddUserToGroupAtomic(tx, usr.ID, usrGroup.ID); err != nil {
			return errors.Newf("add user to group: %v", err)
		}
		if err := f(tx, usr); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	usr.Groups = []Group{*usrGroup}
	usr.Type = *ut

	return usr, nil
}

func (a *Authentication) genPasswordWithHash() (password []byte, passwordH []byte, err error) {
	password, err = a.passGen.SecureRandomBytes(genPassLen)
	if err != nil {
		return nil, nil, errors.Newf("generate password: %v", err)
	}
	passwordH, err = hash(password)
	return
}

func (a *Authentication) genAndInsertToken(tx *sql.Tx, action, loginType, forUsrID, loginID string) ([]byte, error) {
	dbt, err := a.urlTokenGen.SecureRandomBytes(56)
	if err != nil {
		return nil, errors.Newf("generate %s verification db token", loginType)
	}
	if err := a.hashAndInsertToken(tx, action, loginType, forUsrID, loginID, dbt); err != nil {
		return nil, errors.Newf("%s verification db token: %v", loginType, err)
	}
	return dbt, nil
}

func (a *Authentication) genAndInsertCode(tx *sql.Tx, action, loginType, forUsrID, loginID string) ([]byte, error) {
	code, err := a.numGen.SecureRandomBytes(6)
	if err != nil {
		return nil, errors.Newf("generate %s verification code", loginType)
	}
	if err := a.hashAndInsertToken(tx, action, loginType, forUsrID, loginID, code); err != nil {
		return nil, errors.Newf("%s verification code: %v", loginType, err)
	}
	return code, nil
}

func (a *Authentication) hashAndInsertToken(tx *sql.Tx, action, loginType, forUsrID, loginID string, code []byte) error {
	codeH, err := hash([]byte(code))
	if err != nil {
		return errors.Newf("hash for storage: %v", err)
	}

	var insFuncAtomic func(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	var insFunc func(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)

	switch loginType {
	case loginTypePhone:
		insFunc = a.db.InsertPhoneToken
		insFuncAtomic = a.db.InsertPhoneTokenAtomic
	case loginTypeEmail:
		insFunc = a.db.InsertEmailToken
		insFuncAtomic = a.db.InsertEmailTokenAtomic
	default:
		return errors.NewClientf(LoginTypeNotSupportedErrorF, loginType)
	}

	var validity time.Duration

	switch action {
	case ActionResetPass:
		validity = resetValidity
	case ActionVerify:
		validity = verifyValidity
	case ActionInvite:
		validity = inviteValidity
	default:
		return errors.NewClientf(ActionNotSupportedErrorF, action)
	}

	expiry := time.Now().Add(validity)
	if tx == nil {
		_, err = insFunc(forUsrID, loginID, codeH, false, expiry)
	} else {
		_, err = insFuncAtomic(tx, forUsrID, loginID, codeH, false, expiry)
	}
	if err != nil {
		return errors.Newf("insert: %v", err)
	}

	return nil
}

func (a *Authentication) genURL(action, loginType, userID string, dbt []byte) (string, error) {
	URL, err := url.Parse(a.webAppURL)
	if err != nil {
		return "", errors.Newf("parsing web app URL: %v", err)
	}
	URL.Path = path.Join(URL.Path, action, loginType, userID, string(dbt))
	return URL.String(), nil
}

func (a *Authentication) sendEmail(toAddr, subj string, t *template.Template, data interface{}) error {
	emailBf := bytes.NewBuffer(make([]byte, 0, 256))
	err := t.Execute(emailBf, data)
	if err != nil {
		return errors.Newf("email body from template: %v", err)
	}
	err = a.mailer.SendEmail(SendMail{
		ToEmails: []string{toAddr},
		Subject:  subj,
		Body:     template.HTML(emailBf.String()),
	})
	if err != nil {
		return errors.Newf("send email: %v", err)
	}
	return nil
}

func (a *Authentication) sendSMS(toPhone string, t *template.Template, data interface{}) error {
	SMSBf := bytes.NewBuffer(make([]byte, 0, 256))
	err := t.Execute(SMSBf, data)
	if err != nil {
		return errors.Newf("SMS from template: %v", err)
	}
	if err := a.smser.SMS(toPhone, SMSBf.String()); err != nil {
		return errors.Newf("send SMS: %v", err)
	}
	return nil
}

func (a *Authentication) dbTokenValid(userID, dbtStr string,
	f func(userID string, offset, count int64) ([]*DBToken, error)) (*DBToken, error) {

	if dbtStr == "" {
		return nil, errors.NewUnauthorized("confirmation token cannot be empty")
	}
	if userID == "" {
		return nil, errors.NewUnauthorized("userID cannot be empty")
	}
	offset := int64(0)
	count := int64(100)
	var dbt *DBToken
resumeFunc:
	for {
		codes, err := f(userID, offset, count)
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewForbidden("token is invalid")
		}
		if err != nil {
			return nil, errors.Newf("get phone db tokens: %v", err)
		}
		offset = offset + count
		for _, dbt = range codes {
			err = bcrypt.CompareHashAndPassword(dbt.Token, []byte(dbtStr))
			if err != nil {
				continue
			}
			break resumeFunc
		}
	}
	if dbt.IsUsed {
		return nil, errors.NewForbiddenf("token already used")
	}
	if time.Now().After(dbt.ExpiryDate) {
		return nil, errors.NewAuth("token has expired")
	}
	return dbt, nil
}

func (a *Authentication) insertPhoneAtomic(tx *sql.Tx, usr *User, number string) error {
	if number == "" {
		return errors.NewClient("phone number cannot be empty")
	}
	phone, err := a.db.InsertUserPhoneAtomic(tx, usr.ID, number, false)
	if err != nil {
		return errors.Newf("insert phone: %v", err)
	}
	usr.Phone = *phone
	return nil
}

func (a *Authentication) insertEmailAtomic(tx *sql.Tx, usr *User, address string) error {
	if address == "" {
		return errors.NewClient("email address cannot be empty")
	}
	email, err := a.db.InsertUserEmailAtomic(tx, usr.ID, address, false)
	if err != nil {
		return errors.Newf("insert email: %v", err)
	}
	usr.Email = *email
	return nil
}

func (a *Authentication) getOrCreateGroup(groupName string, acl int) (*Group, error) {
	grp, err := a.db.GroupByName(groupName)
	if err != nil {
		if !a.db.IsNotFoundError(err) {
			return nil, errors.Newf("get group by name: %v", err)
		}
		grp, err = a.db.InsertGroup(groupName, acl)
		if err != nil {
			return nil, errors.Newf("insert group: %v", err)
		}
	}
	return grp, nil
}

func (a *Authentication) getOrCreateUserType(name string) (*UserType, error) {
	ut, err := a.db.UserTypeByName(name)
	if err != nil {
		if !a.db.IsNotFoundError(err) {
			return nil, errors.Newf("get user type by name: %v", err)
		}
		ut, err = a.db.InsertUserType(name)
		if err != nil {
			return nil, errors.Newf("insert user type: %v", err)
		}
	}
	return ut, nil
}

func (a *Authentication) validateJWTInGroup(jwt string, g Group) (Claim, error) {
	clm := Claim{}
	if jwt == "" {
		return clm, errors.NewUnauthorizedf("jwt must be provided")
	}
	_, err := a.jwter.Validate(jwt, &clm)
	if err != nil {
		return clm, errors.NewForbidden("invalid jwt")
	}
	if !inGroups(g, clm.Groups) {
		return clm, errors.NewForbidden("invalid jwt")
	}
	return clm, nil
}

func (a *Authentication) loginFacebook(fbToken string) (*User, error) {
	if a.fb == nil {
		return nil, errors.NewNotImplementedf("facebook registration not available")
	}
	fbID, err := a.validateFbToken(fbToken)
	if err != nil {
		return nil, err
	}
	usr, err := a.db.UserByFacebook(fbID)
	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewNotFound("facebook user not registered")
		}
		return nil, errors.Newf("get user by facebook ID: %v", err)
	}
	return usr, nil
}

func (a *Authentication) validateFbToken(fbToken string) (string, error) {
	oa, err := a.fb.ValidateToken(fbToken)
	if err != nil {
		if a.fb.IsAuthError(err) {
			return "", errors.NewAuthf("facebook: %v", err)
		}
		return "", errors.Newf("validate facebook token: %v", err)
	}
	return oa.UserID(), nil
}

func passwordValid(hashed, password []byte) error {
	if err := bcrypt.CompareHashAndPassword(hashed, password); err != nil {
		return errors.NewForbiddenf("invalid username/password combination")
	}
	return nil
}

func hash(password []byte) ([]byte, error) {
	passH, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Newf("hash password: %v", err)
	}
	return passH, nil
}

func hashIfValid(password string) ([]byte, error) {
	if len(password) < minPassLen {
		return nil, errors.NewClientf("password must be at least %d characters", minPassLen)
	}
	return hash([]byte(password))
}

func inGroups(needle Group, haystack []Group) bool {
	for _, straw := range haystack {
		if straw.ID == needle.ID {
			return true
		}
	}
	return false
}

func inStrs(needle string, haystack []string) bool {
	for _, straw := range haystack {
		if straw == needle {
			return true
		}
	}
	return false
}

func formatPhone(phone string) string {
	parts := rePhone.FindAllString(phone, -1)
	formatted := ""
	if strings.HasPrefix(phone, "+") {
		formatted = "+"
	}
	for _, part := range parts {
		formatted = formatted + part
	}
	return formatted
}

func obfuscatePhone(num string) string {
	n := len(num)
	if n < 4 {
		return "xxx"
	}
	showFrom := n - 3
	if n < 6 {
		showFrom = n - 2
	}
	mask := strings.Repeat("x", showFrom)
	return mask + num[showFrom:]
}

func obfuscateEmail(addr string) string {
	j := strings.Index(addr, "@")
	if j == -1 {
		return "xxx"
	}
	num := addr[0:j]
	n := len(num)
	if n < 5 {
		return "xxx" + addr[j:]
	}
	showFrom := n - 3
	if n < 7 {
		showFrom = n - 2
	}
	mask := strings.Repeat("x", showFrom-1)
	return num[0:1] + mask + num[showFrom:] + addr[j:]
}
