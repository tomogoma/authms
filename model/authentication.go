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
	"github.com/tomogoma/go-commons/errors"
	"golang.org/x/crypto/bcrypt"
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

	UpdateUsername(userID, username string) (*Username, error)
	UpdateUserPhone(userID, phone string, verified bool) (*VerifLogin, error)
	UpdateUserEmail(userID, email string, verified bool) (*VerifLogin, error)

	UpdatePassword(userID string, password []byte) error
	UpdatePasswordAtomic(tx *sql.Tx, userID string, password []byte) error
	UpdateUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*VerifLogin, error)
	UpdateUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*VerifLogin, error)

	InsertUserPhone(userID, phone string, verified bool) (*VerifLogin, error)
	InsertUserEmail(userID, email string, verified bool) (*VerifLogin, error)
	InsertUserName(userID, username string) (*Username, error)
	InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*VerifLogin, error)
	InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*VerifLogin, error)
	InsertUserNameAtomic(tx *sql.Tx, userID, username string) (*Username, error)
	InsertUserFacebookIDAtomic(tx *sql.Tx, userID, fbID string, verified bool) (*Facebook, error)
	InsertUserDeviceAtomic(tx *sql.Tx, userID, devID string) (*Device, error)

	InsertPhoneToken(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	InsertEmailToken(userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	InsertPhoneTokenAtomic(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	InsertEmailTokenAtomic(tx *sql.Tx, userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	PhoneTokens(userID string, offset, count int64) ([]*DBToken, error)
	EmailTokens(userID string, offset, count int64) ([]*DBToken, error)

	User(id string) (*User, []byte, error)
	UserByPhone(phone string) (*User, []byte, error)
	UserByEmail(email string) (*User, []byte, error)
	UserByUsername(username string) (*User, []byte, error)
	UserByFacebook(facebookID string) (*User, error)
	UserByDeviceID(devID string) (*User, []byte, error)
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
	ValidateToken(string) (string, error)
}

type SMSer interface {
	SMS(toPhone, message string) error
}

type JWTEr interface {
	Generate(claims jwt.Claims) (string, error)
	Validate(JWT string, claims jwt.Claims) (*jwt.Token, error)
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

// Authentication has the methods for performing auth. Use NewAuthentication()
// to construct.
type Authentication struct {
	// mandatory parameters
	db            AuthStore
	guard         Guard
	jwter         JWTEr
	passGen       SecureRandomByteser
	numGen        SecureRandomByteser
	urlTokenGen   SecureRandomByteser
	lockDevToUser bool
	allowSelfReg  bool
	// optional parameters
	appNameEmptyable     string
	fbNilable            FacebookCl
	smserNilable         SMSer
	mailerNilable        Mailer
	webAppURLEmptyable   string
	invSubjEmptyable     string
	verSubjEmptyable     string
	resPassSubjEmptyable string
	// tail values optional depending on need/type for communication
	loginTpActionTplts map[string]map[string]*template.Template
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

	inviteValidity    = 24 * 30 * time.Hour
	resetValidity     = 2 * time.Hour
	verifyValidity    = 5 * time.Minute
	extendTknValidity = 2 * time.Hour
	tokenValidity     = 1 * time.Hour

	ActionInvite    = "invite"
	ActionVerify    = "verify"
	ActionResetPass = "reset/password"
	ActionExtendTkn = "extend/token"

	loginTypeUsername = "usernames"
	loginTypeEmail    = "phones"
	loginTypePhone    = "emails"
	loginTypeFacebook = "facebook"
)

var (
	rePhone        = regexp.MustCompile(rePhoneChars)
	validUserTypes = []string{UserTypeIndividual, UserTypeCompany}

	actionNotSupportedErrorF    = "action not supported for request: %s"
	loginTypeNotSupportedErrorF = "login type not supported for request: %s"

	errorBadCreds      = errors.NewUnauthorized("invalid credentials")
	errorNoneDeviceReg = errors.NewForbidden("registration closed to the public unless from accepted device")
	errorFbNotAvail    = errors.NewNotImplementedf("facebook registration not available")
)

// NewAuthentication constructs an Authentication structs or returns an error
// if invalid parameters were provided.
func NewAuthentication(db AuthStore, g Guard, j JWTEr, opts ...Option) (*Authentication, error) {
	if db == nil {
		return nil, errors.New("AuthStore was nil")
	}
	if g == nil {
		return nil, errors.New("Guard was nil")
	}
	if j == nil {
		return nil, errors.New("JWTEr was nil")
	}

	c := &authenticationConfig{}
	c.initializeValues()
	var err error
	if err = c.assignOptions(opts); err != nil {
		return nil, err
	}
	if err = c.fillDefaults(); err != nil {
		return nil, err
	}
	if err = c.valid(); err != nil {
		return nil, err
	}

	return &Authentication{
		db:                   db,
		guard:                g,
		jwter:                j,
		passGen:              c.passGen,
		numGen:               c.numGen,
		urlTokenGen:          c.urlTokenGen,
		allowSelfReg:         c.allowSelfReg,
		lockDevToUser:        c.lockDevToUser,
		appNameEmptyable:     c.appNameEmptyable,
		fbNilable:            c.fbNilable,
		smserNilable:         c.smserNilable,
		mailerNilable:        c.mailerNilable,
		webAppURLEmptyable:   c.webAppURLEmptyable,
		invSubjEmptyable:     c.invSubjEmptyable,
		verSubjEmptyable:     c.verSubjEmptyable,
		resPassSubjEmptyable: c.resPassSubjEmptyable,
		loginTpActionTplts:   c.loginTpActionTplts,
	}, nil
}

// RegisterSelfByUsername registers a new user account using username/password combination.
func (a *Authentication) RegisterSelfByUsername(clID, apiKey, userType, username, password string) (*User, error) {
	if a.lockDevToUser {
		return nil, errorNoneDeviceReg
	}
	return a.registerSelf(clID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		if username == "" {
			return errors.NewClient("username cannot be empty")
		}
		_, _, err := a.db.UserByUsername(username)
		if err = a.usrIdentifierAvail(loginTypeUsername, err); err != nil {
			return err
		}
		uname, err := a.db.InsertUserNameAtomic(tx, usr.ID, username)
		if err != nil {
			return errors.Newf("insert username: %v", err)
		}
		usr.UserName = *uname
		return nil
	})
}

// RegisterSelfByLockedPhone registers a new user account using phone/deviceID/password combination.
func (a *Authentication) RegisterSelfByLockedPhone(clID, apiKey, userType, devID, number, password string) (*User, error) {
	return a.registerSelf(clID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		_, _, err := a.db.UserByDeviceID(devID)
		if err := a.usrIdentifierAvail(loginTypePhone, err); err != nil {
			return err
		}
		dev, err := a.db.InsertUserDeviceAtomic(tx, usr.ID, devID)
		if err != nil {
			return errors.Newf("insert device: %v", err)
		}
		usr.Devices = []Device{*dev}
		if err := a.insertPhoneAtomic(tx, usr, number); err != nil {
			return err
		}
		if a.smserNilable == nil {
			return nil
		}
		_, err = a.genAndSendTokens(tx, ActionVerify, loginTypePhone, number, usr.ID)
		return err
	})
}

// RegisterSelfByPhone registers a new user account using phone/password combination.
func (a *Authentication) RegisterSelfByPhone(clID, apiKey, userType, number, password string) (*User, error) {
	if a.lockDevToUser {
		return nil, errorNoneDeviceReg
	}
	return a.registerSelf(clID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		if err := a.insertPhoneAtomic(tx, usr, number); err != nil {
			return err
		}
		if a.smserNilable == nil {
			return nil
		}
		_, err := a.genAndSendTokens(tx, ActionVerify, loginTypePhone, number, usr.ID)
		return err
	})
}

// RegisterSelfByEmail registers a new user account using email/password combination.
func (a *Authentication) RegisterSelfByEmail(clID, apiKey, userType, address, password string) (*User, error) {
	if a.lockDevToUser {
		return nil, errorNoneDeviceReg
	}
	return a.registerSelf(clID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		if err := a.insertEmailAtomic(tx, usr, address); err != nil {
			return err
		}
		if a.mailerNilable == nil {
			return nil
		}
		_, err := a.genAndSendTokens(tx, ActionVerify, loginTypeEmail, address, usr.ID)
		return err
	})
}

// RegisterSelfByFacebook registers a new user account using facebook OAuth2 for auth.
func (a *Authentication) RegisterSelfByFacebook(clID, apiKey, userType, fbToken string) (*User, error) {
	if a.lockDevToUser {
		return nil, errorNoneDeviceReg
	}
	if a.fbNilable == nil {
		return nil, errorFbNotAvail
	}
	passwordB, err := a.passGen.SecureRandomBytes(genPassLen)
	if err != nil {
		return nil, errors.Newf("generate password: %v", err)
	}
	return a.registerSelf(clID, apiKey, userType, string(passwordB), func(tx *sql.Tx, usr *User) error {
		if fbToken == "" {
			return errors.NewClient("facebook token cannot be empty")
		}
		fbID, err := a.validateFbToken(fbToken)
		if err != nil {
			return err
		}
		_, err = a.db.UserByFacebook(fbID)
		if err = a.usrIdentifierAvail(loginTypeFacebook, err); err != nil {
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

// CreateByPhone registers an account on behalf of a new user. The new user will
// receive a phone invitation to complete account creation using SetPassword().
func (a *Authentication) CreateByPhone(clientID, apiKey, JWT, userType, number, groupID string) (*User, error) {
	if a.smserNilable == nil {
		return nil, errors.NewNotImplementedf("SMS notification (to created user) not available")
	}
	return a.createUser(clientID, apiKey, JWT, userType, groupID, func(tx *sql.Tx, usr *User) error {
		if err := a.insertPhoneAtomic(tx, usr, number); err != nil {
			return err
		}
		_, err := a.genAndSendTokens(tx, ActionInvite, loginTypePhone, number, usr.ID)
		return err
	})
}

// CreateByEmail registers an account on behalf of a new user. The new user will
// receive an email invitation to complete account creation using SetPassword().
func (a *Authentication) CreateByEmail(clientID, apiKey, JWT, userType, address, groupID string) (*User, error) {
	if a.mailerNilable == nil {
		return nil, errors.NewNotImplementedf("email notification (to created user) not available")
	}
	return a.createUser(clientID, apiKey, JWT, userType, groupID, func(tx *sql.Tx, usr *User) error {
		if err := a.insertEmailAtomic(tx, usr, address); err != nil {
			return err
		}
		_, err := a.genAndSendTokens(tx, ActionInvite, loginTypeEmail, address, usr.ID)
		return err
	})
}

// UpdateUsername updates a user account's username.
func (a *Authentication) UpdateUsername(clientID, apiKey, JWT, newUsrName string) (*Username, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	// TODO allow admin to update
	clm := new(JWTClaim)
	if _, err := a.jwter.Validate(JWT, clm); err != nil {
		return nil, err
	}
	_, _, err := a.db.UserByUsername(newUsrName)
	if err = a.usrIdentifierAvail(loginTypeUsername, err); err != nil {
		return nil, err
	}
	uname, err := a.db.InsertUserName(clm.UsrID, newUsrName)
	if err != nil {
		return nil, errors.Newf("user by phone: %v", err)
	}
	return uname, nil
}

// UpdateUsername updates a user account's phone.
func (a *Authentication) UpdatePhone(clientID, apiKey, JWT, newNum string) (*VerifLogin, error) {

	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	// TODO allow admin to update
	clm := new(JWTClaim)
	if _, err := a.jwter.Validate(JWT, clm); err != nil {
		return nil, err
	}
	newNum, err := formatValidPhone(newNum)
	if err != nil {
		return nil, err
	}
	_, _, err = a.db.UserByPhone(newNum)
	if err = a.usrIdentifierAvail(loginTypePhone, err); err != nil {
		return nil, err
	}
	phone, err := a.db.InsertUserPhone(clm.UsrID, newNum, false)
	if err != nil {
		return nil, errors.Newf("update phone: %v", err)
	}
	if a.smserNilable == nil {
		return phone, nil
	}
	_, err = a.genAndSendTokens(nil, ActionVerify, loginTypePhone, newNum, clm.UsrID)
	if err != nil {
		return nil, err
	}
	return phone, nil
}

// UpdatePassword updates a user account's password.
func (a *Authentication) UpdatePassword(clientID, apiKey, JWT, old, new string) error {

	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return err
	}

	clm := new(JWTClaim)
	if _, err := a.jwter.Validate(JWT, clm); err != nil {
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

// SetPassword updates a user account's password following a SendPassResetCode()
// request. dbt is the token initially sent to the user for verification.
// loginType should be similar to the one used during SendPassResetCode().
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
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
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
		addr, err = updtVerifiedFunc(tx, userID, tkn.Address, true)
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

// SendVerCode sends a verification code to toAddr to verify the
// address. loginType determines determines whether toAddr is a phone or an email.
// subsequent calls to VerifyDBT() or VerifyAndExtendDBT() with the correct code
// completes the verification.
func (a *Authentication) SendVerCode(clientID, apiKey, loginType, JWT, toAddr string) (string, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return "", err
	}

	clms := new(JWTClaim)
	if _, err := a.jwter.Validate(JWT, clms); err != nil {
		return "", err
	}

	var err error
	var usr *User
	var insFunc func(userID, phone string, verified bool) (*VerifLogin, error)
	var isMessengerAvail bool

	switch loginType {
	case loginTypePhone:
		insFunc = a.db.InsertUserPhone
		usr, _, err = a.db.UserByPhone(toAddr)
		isMessengerAvail = a.smserNilable != nil
	case loginTypeEmail:
		insFunc = a.db.InsertUserEmail
		usr, _, err = a.db.UserByEmail(toAddr)
		isMessengerAvail = a.mailerNilable != nil
	default:
		return "", errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	if !isMessengerAvail {
		return "", errors.NewNotImplementedf("notification method not available for %s", loginType)
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

// SendPassResetCode sends a password reset code to toAddr to allow a user
// to reset their forgotten password.
// loginType determines whether toAddr is a phone or an email.
// subsequent calls to SetPassword() with the correct code completes the
// password reset.
func (a *Authentication) SendPassResetCode(clientID, apiKey, loginType, toAddr string) (string, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return "", err
	}
	var err error
	var usr *User
	var isMessengerAvail bool

	switch loginType {
	case loginTypePhone:
		usr, _, err = a.db.UserByPhone(toAddr)
		isMessengerAvail = a.smserNilable != nil
	case loginTypeEmail:
		usr, _, err = a.db.UserByEmail(toAddr)
		isMessengerAvail = a.mailerNilable != nil
	default:
		return "", errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	if !isMessengerAvail {
		return "", errors.NewNotImplementedf("notification method not available for %s", loginType)
	}

	if err != nil {
		if a.db.IsNotFoundError(err) {
			return "", errors.NewForbiddenf("%s does not exist", toAddr)
		}
		return "", errors.Newf("user by %s: %v", loginType, err)
	}

	return a.genAndSendTokens(nil, ActionResetPass, loginType, toAddr, usr.ID)
}

// VerifyAndExtendDBT verifies a user's address and returns a temporary token
// that can be used to perform actions that would otherwise not be possible on
// the user's account without a password or a JWT for a limited period of time.
// See VerifyDBT() for details on verification.
func (a *Authentication) VerifyAndExtendDBT(clID, apiKey, lt, usrID, dbt string) (string, error) {
	if err := a.guard.APIKeyValid(clID, apiKey); err != nil {
		return "", err
	}
	lv, err := a.verifyDBT(lt, usrID, dbt)
	if err != nil {
		return "", err
	}
	tkn, err := a.genAndInsertToken(nil, ActionExtendTkn, lt, usrID, lv.Address)
	if err != nil {
		return "", err
	}
	return string(tkn), nil
}

// VerifyDBT sets a user's address as verified after successful SendVerCode()
// and subsequent entry of the code by the user.
// loginType should be similar to the one used during SendVerCode().
func (a *Authentication) VerifyDBT(clientID, apiKey, loginType, userID, dbt string) (*VerifLogin, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	return a.verifyDBT(loginType, userID, dbt)
}

// Login validates a user's credentials and returns the user's information
// together with a JWT for subsequent requests to this and other micro-services.
func (a *Authentication) Login(clientID, apiKey, loginType, identifier, password string) (*User, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}

	var usr *User
	var passHB []byte
	var err error

	switch loginType {
	case loginTypePhone:
		identifier, err := formatValidPhone(identifier)
		if err != nil {
			return nil, errorBadCreds
		}
		usr, passHB, err = a.db.UserByPhone(identifier)
	case loginTypeEmail:
		usr, passHB, err = a.db.UserByEmail(identifier)
	case loginTypeUsername:
		usr, passHB, err = a.db.UserByUsername(identifier)
	case loginTypeFacebook:
		return a.loginFacebook(identifier)
	default:
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}
	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errorBadCreds
		}
		return nil, errors.Newf("get user by %s: %v", loginType, err)
	}

	if err := passwordValid(passHB, []byte(password)); err != nil {
		return nil, err
	}

	usr.JWT, err = a.jwter.Generate(newJWTClaim(usr.ID, usr.Groups))
	if err != nil {
		return nil, errors.Newf("generate JWT: %v", err)
	}

	return usr, nil
}

func (a *Authentication) usrIdentifierAvail(loginType string, fetchErr error) error {
	if fetchErr == nil {
		return errors.NewClientf("%s not available", loginType)
	}
	if !a.db.IsNotFoundError(fetchErr) {
		return errors.Newf("user by %s: %v", loginType, fetchErr)
	}
	return nil
}

func (a *Authentication) verifyDBT(loginType, userID, dbt string) (*VerifLogin, error) {

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
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	tkn, err := a.dbTokenValid(userID, dbt, tokensFetchFunc)
	if err != nil {
		return nil, err
	}
	phone, err := updateLoginFunc(userID, tkn.Address, true)
	if err != nil {
		return nil, errors.Newf("update phone to verified: %v", err)
	}
	return phone, nil
}

func (a *Authentication) genAndSendTokens(tx *sql.Tx, action, loginType, toAddr, usrID string) (string, error) {

	URL := ""
	if a.webAppURLEmptyable != "" {
		tkn, err := a.genAndInsertToken(tx, action, loginType, usrID, toAddr)
		if err != nil {
			return "", err
		}
		URL, err = a.genURL(action, loginType, usrID, tkn)
		if err != nil {
			return "", err
		}
	}

	code, err := a.genAndInsertCode(tx, action, loginType, usrID, toAddr)
	if err != nil {
		return "", err
	}

	var sendData interface{}
	var subj string

	switch action {
	case ActionInvite:
		sendData = InvitationTemplate{AppName: a.appNameEmptyable, URLToken: URL}
		subj = a.invSubjEmptyable
	case ActionVerify:
		sendData = VerificationTemplate{AppName: a.appNameEmptyable, URLToken: URL,
			Code: string(code)}
		subj = a.verSubjEmptyable
	case ActionResetPass:
		sendData = VerificationTemplate{AppName: a.appNameEmptyable, URLToken: URL,
			Code: string(code)}
		subj = a.resPassSubjEmptyable
	default:
		return "", errors.Newf(actionNotSupportedErrorF, action)
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
		return "", errors.Newf(loginTypeNotSupportedErrorF, loginType)
	}
	if err != nil {
		return "", err
	}

	return obfuscateFunc(toAddr), nil
}

func (a *Authentication) registerSelf(clID, apiKey, userType,
	password string, f func(tx *sql.Tx, usr *User) error) (*User, error) {

	if err := a.guard.APIKeyValid(clID, apiKey); err != nil {
		return nil, err
	}
	if !a.allowSelfReg {
		return nil, errors.NewForbidden("registration closed from the public")
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

func (a *Authentication) createUser(clientID, apiKey, JWT, userType,
	groupID string, f func(tx *sql.Tx, usr *User) error) (*User, error) {

	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	adminGrp, err := a.getOrCreateGroup(GroupAdmin, AccessLevelAdmin)
	if err != nil {
		return nil, err
	}
	if _, err := a.validateJWTInGroup(JWT, *adminGrp); err != nil {
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
		return errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	var validity time.Duration

	switch action {
	case ActionResetPass:
		validity = resetValidity
	case ActionVerify:
		validity = verifyValidity
	case ActionInvite:
		validity = inviteValidity
	case ActionExtendTkn:
		validity = extendTknValidity
	default:
		return errors.NewClientf(actionNotSupportedErrorF, action)
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
	URL, err := url.Parse(a.webAppURLEmptyable)
	if err != nil {
		return "", errors.Newf("parsing web app URL: %v", err)
	}
	URL.Path = path.Join(URL.Path, action, loginType, userID, string(dbt))
	return URL.String(), nil
}

func (a *Authentication) sendEmail(toAddr, subj string, t *template.Template, data interface{}) error {
	if a.mailerNilable == nil {
		return errors.New("Mailer was nil")
	}
	emailBf := bytes.NewBuffer(make([]byte, 0, 256))
	err := t.Execute(emailBf, data)
	if err != nil {
		return errors.Newf("email body from template: %v", err)
	}
	err = a.mailerNilable.SendEmail(SendMail{
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
	if a.smserNilable == nil {
		return errors.New("SMSer was nil")
	}
	SMSBf := bytes.NewBuffer(make([]byte, 0, 256))
	err := t.Execute(SMSBf, data)
	if err != nil {
		return errors.Newf("SMS from template: %v", err)
	}
	if err := a.smserNilable.SMS(toPhone, SMSBf.String()); err != nil {
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
	number, err := formatValidPhone(number)
	if err != nil {
		return err
	}
	_, _, err = a.db.UserByPhone(number)
	if err = a.usrIdentifierAvail(loginTypePhone, err); err != nil {
		return err
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
	_, _, err := a.db.UserByEmail(address)
	if err = a.usrIdentifierAvail(loginTypeEmail, err); err != nil {
		return err
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

func (a *Authentication) validateJWTInGroup(JWT string, g Group) (JWTClaim, error) {
	clm := JWTClaim{}
	if JWT == "" {
		return clm, errors.NewUnauthorizedf("JWT must be provided")
	}
	_, err := a.jwter.Validate(JWT, &clm)
	if err != nil {
		return clm, errors.NewForbidden("invalid JWT")
	}
	if !inGroups(g, clm.Groups) {
		return clm, errors.NewForbidden("invalid JWT")
	}
	return clm, nil
}

func (a *Authentication) loginFacebook(fbToken string) (*User, error) {
	if a.fbNilable == nil {
		return nil, errorFbNotAvail
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
	if a.fbNilable == nil {
		return "", errorFbNotAvail
	}
	fbUsrID, err := a.fbNilable.ValidateToken(fbToken)
	if err != nil {
		if a.fbNilable.IsAuthError(err) {
			return "", errors.NewAuthf("facebook: %v", err)
		}
		return "", errors.Newf("validate facebook token: %v", err)
	}
	return fbUsrID, nil
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

func formatValidPhone(number string) (string, error) {
	number = formatPhone(number)
	if number == "" {
		return "", errors.NewClient("phone number was empty")
	}
	return number, nil
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
