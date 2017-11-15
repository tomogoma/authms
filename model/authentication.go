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

	"reflect"

	"github.com/dgrijalva/jwt-go"
	errors "github.com/tomogoma/go-typed-errors"
	"golang.org/x/crypto/bcrypt"
)

type AuthStore interface {
	IsNotFoundError(error) bool
	ExecuteTx(fn func(*sql.Tx) error) error

	InsertGroup(name string, acl float32) (*Group, error)
	Group(string) (*Group, error)
	GroupByName(string) (*Group, error)

	InsertUserType(name string) (*UserType, error)
	UserTypeByName(string) (*UserType, error)

	HasUsers(groupID string) error
	InsertUserAtomic(tx *sql.Tx, t UserType, password []byte) (*User, error)
	UpdatePassword(userID string, password []byte) error
	UpdatePasswordAtomic(tx *sql.Tx, userID string, password []byte) error
	User(id string) (*User, []byte, error)
	UserByDeviceID(devID string) (*User, []byte, error)
	UserByUsername(username string) (*User, []byte, error)
	UserByPhone(phone string) (*User, []byte, error)
	UserByEmail(email string) (*User, []byte, error)
	UserByFacebook(facebookID string) (*User, error)

	AddUserToGroupAtomic(tx *sql.Tx, userID, groupID string) error

	InsertUserDeviceAtomic(tx *sql.Tx, userID, devID string) (*Device, error)

	InsertUserName(userID, username string) (*Username, error)
	InsertUserNameAtomic(tx *sql.Tx, userID, username string) (*Username, error)

	InsertUserPhone(userID, phone string, verified bool) (*VerifLogin, error)
	InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*VerifLogin, error)
	UpdateUserPhone(userID, phone string, verified bool) (*VerifLogin, error)
	UpdateUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*VerifLogin, error)

	InsertPhoneToken(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	InsertPhoneTokenAtomic(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	PhoneTokens(userID string, offset, count int64) ([]DBToken, error)

	InsertUserEmail(userID, email string, verified bool) (*VerifLogin, error)
	InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*VerifLogin, error)
	UpdateUserEmail(userID, email string, verified bool) (*VerifLogin, error)
	UpdateUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*VerifLogin, error)

	InsertEmailToken(userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	InsertEmailTokenAtomic(tx *sql.Tx, userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	EmailTokens(userID string, offset, count int64) ([]DBToken, error)

	InsertUserFbIDAtomic(tx *sql.Tx, userID, fbID string, verified bool) (*Facebook, error)
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
	SendEmail(email SendMail) error
}

// Authentication has the methods for performing auth. Use NewAuthentication()
// to construct.
type Authentication struct {
	errors.AllErrCheck
	// mandatory parameters
	db            AuthStore
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
	webAppURLNilable     *url.URL
	invSubjEmptyable     string
	verSubjEmptyable     string
	resPassSubjEmptyable string
	// tail values optional depending on need/type for communication
	loginTpActionTplts map[string]map[string]*template.Template
}

type regConditions func(id string) (string, error)
type regFunc func(tx *sql.Tx, actionType, id string, usr *User) error

const (
	GroupSuper  = "super"
	GroupAdmin  = "admin"
	GroupStaff  = "staff"
	GroupPublic = "public"

	AccessLevelSuper  = 1
	AccessLevelAdmin  = 3
	AccessLevelStaff  = 7
	AccessLevelPublic = 9

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

	LoginTypeUsername = "usernames"
	LoginTypeEmail    = "emails"
	LoginTypePhone    = "phones"
	LoginTypeFacebook = "facebook"
	LoginTypeDev      = "devices"
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
func NewAuthentication(db AuthStore, j JWTEr, opts ...Option) (*Authentication, error) {
	if db == nil || reflect.ValueOf(db).IsNil() {
		return nil, errors.New("AuthStore was nil")
	}
	if j == nil || reflect.ValueOf(j).IsNil() {
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
		webAppURLNilable:     c.webAppURLNilable,
		invSubjEmptyable:     c.invSubjEmptyable,
		verSubjEmptyable:     c.verSubjEmptyable,
		resPassSubjEmptyable: c.resPassSubjEmptyable,
		loginTpActionTplts:   c.loginTpActionTplts,
	}, nil
}

func (a *Authentication) CanRegisterFirst() (bool, error) {

	superGrp, err := a.getOrCreateGroup(GroupSuper, AccessLevelSuper)
	if err != nil {
		return false, err
	}

	err = a.db.HasUsers(superGrp.ID)
	if err == nil {
		return false, nil
	}
	if a.db.IsNotFoundError(err) {
		return true, nil
	}

	return false, errors.Newf("check db has users: %v", err)
}

func (a *Authentication) RegisterFirst(loginType, userType, id string, secret []byte) (*User, error) {

	ok, err := a.CanRegisterFirst()
	if err != nil {
		return nil, errors.Newf("check ok to register first user: %v", err)
	}
	if !ok {
		return nil, errors.NewForbidden("Nothing to see here")
	}

	var regF regFunc
	var regCondF regConditions
	switch loginType {
	case LoginTypeUsername:
		regCondF = a.regUsernameConditions
		regF = a.regUsername
	case LoginTypeEmail:
		regCondF = a.regEmailConditions
		regF = a.regEmail
	case LoginTypePhone:
		regCondF = a.regPhoneConditions
		regF = a.regPhone
	case LoginTypeFacebook:
		if a.fbNilable == nil {
			return nil, errorFbNotAvail
		}
		var err error
		secret, err = a.passGen.SecureRandomBytes(genPassLen)
		if err != nil {
			return nil, errors.Newf("generate password: %v", err)
		}
		regCondF = a.regFacebookConditions
		regF = a.regFacebook
	default:
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	superGrp, err := a.getOrCreateGroup(GroupSuper, AccessLevelSuper)
	if err != nil {
		return nil, err
	}

	// Assume system is a 'person' in group super registering another
	// of the same group - the first human 'person'.
	// This bypasses restrictions on registerSelf() e.g.
	// 1. User can only be a member of the public group.
	// 2. Self registration may be disabled by config options.
	return a.registerOther(*superGrp, userType, superGrp.ID, id, secret, regCondF, regF, ActionVerify)
}

// RegisterSelf registers a new user account using id secret combination.
// It should not be possible to register using this method if WithDevLockedToUser()
// was given a true value.
func (a *Authentication) RegisterSelf(loginType, userType, id string, secret []byte) (*User, error) {

	if a.lockDevToUser {
		return nil, errorNoneDeviceReg
	}

	var regF regFunc
	var regCondF regConditions
	switch loginType {
	case LoginTypeUsername:
		regCondF = a.regUsernameConditions
		regF = a.regUsername
	case LoginTypeEmail:
		regCondF = a.regEmailConditions
		regF = a.regEmail
	case LoginTypePhone:
		regCondF = a.regPhoneConditions
		regF = a.regPhone
	case LoginTypeFacebook:
		if a.fbNilable == nil {
			return nil, errorFbNotAvail
		}
		var err error
		secret, err = a.passGen.SecureRandomBytes(genPassLen)
		if err != nil {
			return nil, errors.Newf("generate password: %v", err)
		}
		regCondF = a.regFacebookConditions
		regF = a.regFacebook
	default:
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	return a.registerSelf(userType, id, secret, regCondF, regF)
}

// RegisterSelfByLockedPhone registers a new user account using phone/deviceID/password combination.
func (a *Authentication) RegisterSelfByLockedPhone(userType, devID, number string, password []byte) (*User, error) {
	return a.registerSelf(userType, number, password,
		func(number string) (string, error) {
			if _, err := a.regDevConditions(devID); err != nil {
				return "", err
			}
			return a.regPhoneConditions(number)
		},
		func(tx *sql.Tx, actionType, number string, usr *User) error {
			if err := a.regDevice(tx, actionType, devID, usr); err != nil {
				return err
			}
			return a.regPhone(tx, actionType, number, usr)
		},
	)
}

func (a *Authentication) RegisterOther(JWT, newLoginType, userType, id, groupID string) (*User, error) {
	adminGrp, err := a.getOrCreateGroup(GroupAdmin, AccessLevelAdmin)
	if err != nil {
		return nil, err
	}
	superGroup, err := a.getOrCreateGroup(GroupSuper, AccessLevelSuper)
	if err != nil {
		return nil, err
	}
	clm, err := a.validateJWTInOneOf(JWT, *adminGrp, *superGroup)
	if err != nil {
		return nil, err
	}

	var regCondF regConditions
	var regF regFunc
	switch newLoginType {
	case LoginTypePhone:
		if a.smserNilable == nil {
			return nil, errors.NewNotImplementedf("SMS notification (to created user) not available")
		}
		regCondF = a.regPhoneConditions
		regF = a.regPhone
	case LoginTypeEmail:
		if a.mailerNilable == nil {
			return nil, errors.NewNotImplementedf("email notification (to created user) not available")
		}
		regCondF = a.regEmailConditions
		regF = a.regEmail
	default:
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, newLoginType)
	}

	pass, err := a.passGen.SecureRandomBytes(genPassLen)
	if err != nil {
		return nil, errors.Newf("generate password: %v", err)
	}

	// clm.StrongestGroup cannot panic because we validate that JWT claims
	// to be in either admin or super groups or both.
	return a.registerOther(*clm.StrongestGroup, userType, groupID, id, pass, regCondF, regF)
}

// UpdateIdentifier updates a user account's visible identifier to newID for
// loginType.
func (a *Authentication) UpdateIdentifier(JWT, loginType, newId string) (*User, error) {
	clm := new(JWTClaim)
	if _, err := a.jwter.Validate(JWT, clm); err != nil {
		return nil, err
	}

	usr, _, err := a.db.User(clm.UsrID)
	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errors.Newf("not found user by JWT provided userID: %v", err)
		}
		return nil, errors.Newf("get user: %v", err)
	}

	switch loginType {
	case LoginTypeUsername:
		var usrnm *Username
		usrnm, err = a.updateUsername(usr.ID, newId)
		usr.UserName = *usrnm
	case LoginTypePhone:
		var phn *VerifLogin
		phn, err = a.updatePhone(usr.ID, newId)
		usr.Phone = *phn
	case LoginTypeEmail:
		err = errors.NewNotImplemented()
	case LoginTypeFacebook:
		err = errors.NewNotImplemented()
	default:
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}
	return usr, err
}

// UpdatePassword updates a user account's password.
func (a *Authentication) UpdatePassword(JWT string, old, newPass []byte) error {
	clm := new(JWTClaim)
	if _, err := a.jwter.Validate(JWT, clm); err != nil {
		return err
	}
	_, oldPassH, err := a.db.User(clm.UsrID)
	if err != nil {
		return errors.Newf("get user: %v", err)
	}

	if err = passwordValid(oldPassH, old); err != nil {
		return err
	}
	newPassH, err := hashIfValid(newPass)
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
func (a *Authentication) SetPassword(loginType, forAddr string, dbt, pass []byte) (*VerifLogin, error) {
	var tkn *DBToken
	var err error
	var updtVerifiedFunc func(*sql.Tx, string, string, bool) (*VerifLogin, error)

	var usr *User
	switch loginType {
	case LoginTypeEmail:
		usr, _, err = a.db.UserByEmail(forAddr)
	case LoginTypePhone:
		usr, _, err = a.db.UserByPhone(forAddr)
	default:
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewNotFound("User not found")
		}
		return nil, errors.Newf("fetch user by %s: %v", loginType, err)
	}

	switch loginType {
	case LoginTypeEmail:
		updtVerifiedFunc = a.db.UpdateUserEmailAtomic
		tkn, err = a.dbTokenValid(usr.ID, dbt, a.db.EmailTokens)
	case LoginTypePhone:
		updtVerifiedFunc = a.db.UpdateUserPhoneAtomic
		tkn, err = a.dbTokenValid(usr.ID, dbt, a.db.PhoneTokens)
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
		err = a.db.UpdatePasswordAtomic(tx, usr.ID, passH)
		if err != nil {
			return errors.Newf("update password: %v", err)
		}
		addr, err = updtVerifiedFunc(tx, usr.ID, tkn.Address, true)
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
func (a *Authentication) SendVerCode(JWT, loginType, toAddr string) (*DBTStatus, error) {
	clms := new(JWTClaim)
	if _, err := a.jwter.Validate(JWT, clms); err != nil {
		return nil, err
	}

	var err error
	var usr *User
	var insFunc func(userID, phone string, verified bool) (*VerifLogin, error)
	var isMessengerAvail bool

	switch loginType {
	case LoginTypePhone:
		insFunc = a.db.InsertUserPhone
		usr, _, err = a.db.UserByPhone(toAddr)
		isMessengerAvail = a.smserNilable != nil
	case LoginTypeEmail:
		insFunc = a.db.InsertUserEmail
		usr, _, err = a.db.UserByEmail(toAddr)
		isMessengerAvail = a.mailerNilable != nil
	default:
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	if !isMessengerAvail {
		return nil, errors.NewNotImplementedf("notification method not available for %s", loginType)
	}

	if err != nil {
		if !a.db.IsNotFoundError(err) {
			return nil, errors.Newf("user by %s: %v", loginType, err)
		}

		usr, _, err = a.db.User(clms.UsrID)
		if err != nil {
			if a.db.IsNotFoundError(err) {
				return nil, errors.Newf("valid claims with non-exist user found: %+v", clms)
			}
			return nil, errors.Newf("get user: %v", err)
		}

		_, err = insFunc(usr.ID, toAddr, false)
		if err != nil {
			return nil, errors.Newf("insert toAddr: %v", err)
		}

	} else if usr.ID != clms.UsrID {
		return nil, errors.NewForbiddenf("%s does not belong to you: %v", loginType)
	}

	return a.genAndSendTokens(nil, ActionVerify, loginType, toAddr, usr.ID)

}

// SendPassResetCode sends a password reset code to toAddr to allow a user
// to reset their forgotten password.
// loginType determines whether toAddr is a phone or an email.
// subsequent calls to SetPassword() with the correct code completes the
// password reset.
func (a *Authentication) SendPassResetCode(loginType, toAddr string) (*DBTStatus, error) {
	var err error
	var usr *User
	var isMessengerAvail bool

	switch loginType {
	case LoginTypePhone:
		usr, _, err = a.db.UserByPhone(toAddr)
		isMessengerAvail = a.smserNilable != nil
	case LoginTypeEmail:
		usr, _, err = a.db.UserByEmail(toAddr)
		isMessengerAvail = a.mailerNilable != nil
	default:
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	if !isMessengerAvail {
		return nil, errors.NewNotImplementedf("notification method not available for %s", loginType)
	}

	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewNotFoundf("%s does not exist", toAddr)
		}
		return nil, errors.Newf("user by %s: %v", loginType, err)
	}

	return a.genAndSendTokens(nil, ActionResetPass, loginType, toAddr, usr.ID)
}

// VerifyAndExtendDBT verifies a user's address and returns a temporary token
// that can be used to perform actions that would otherwise not be possible on
// the user's account without a password or a JWT for a limited period of time.
// See VerifyDBT() for details on verification.
func (a *Authentication) VerifyAndExtendDBT(lt, forAddr string, dbt []byte) (string, error) {
	lv, err := a.verifyDBT(lt, forAddr, dbt)
	if err != nil {
		return "", err
	}
	expiry := time.Now().Add(extendTknValidity)
	tkn, err := a.genAndInsertToken(nil, expiry, lt, lv.UserID, lv.Address)
	if err != nil {
		return "", err
	}
	return string(tkn), nil
}

// VerifyDBT sets a user's address as verified after successful SendVerCode()
// and subsequent entry of the code by the user.
// loginType should be similar to the one used during SendVerCode().
func (a *Authentication) VerifyDBT(loginType, forAddr string, dbt []byte) (*VerifLogin, error) {
	return a.verifyDBT(loginType, forAddr, dbt)
}

// Login validates a user's credentials and returns the user's information
// together with a JWT for subsequent requests to this and other micro-services.
func (a *Authentication) Login(loginType, identifier string, password []byte) (*User, error) {
	var usr *User
	var passHB []byte
	var err error

	switch loginType {
	case LoginTypePhone:
		identifier, err := formatValidPhone(identifier)
		if err != nil {
			return nil, errorBadCreds
		}
		usr, passHB, err = a.db.UserByPhone(identifier)
	case LoginTypeEmail:
		usr, passHB, err = a.db.UserByEmail(identifier)
	case LoginTypeUsername:
		usr, passHB, err = a.db.UserByUsername(identifier)
	case LoginTypeFacebook:
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

	if err := passwordValid(passHB, password); err != nil {
		return nil, err
	}

	usr.JWT, err = a.jwter.Generate(newJWTClaim(usr.ID, usr.Groups))
	if err != nil {
		return nil, errors.Newf("generate JWT: %v", err)
	}

	return usr, nil
}

func (a *Authentication) GetUserDetails(JWT string, userID string) (*User, error) {
	clms := new(JWTClaim)
	if _, err := a.jwter.Validate(JWT, clms); err != nil {
		return nil, err
	}
	if clms.UsrID != userID {
		if clms.StrongestGroup == nil {
			return nil, errors.NewForbiddenf("lack sufficient privilege to access this resource")
		}
		if clms.StrongestGroup.AccessLevel > AccessLevelStaff {
			return nil, errors.NewForbiddenf("lack sufficient privilege to access this resource")
		}
	}
	usr, _, err := a.db.User(userID)
	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewNotFound(err)
		}
		return nil, errors.Newf("fetch user: %v", err)
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

func (a *Authentication) verifyDBT(loginType, forAddr string, dbt []byte) (*VerifLogin, error) {

	var tokensFetchFunc func(string, int64, int64) ([]DBToken, error)
	var updateLoginFunc func(string, string, bool) (*VerifLogin, error)
	var usr *User
	var err error

	switch loginType {
	case LoginTypeEmail:
		tokensFetchFunc = a.db.EmailTokens
		updateLoginFunc = a.db.UpdateUserEmail
		usr, _, err = a.db.UserByEmail(forAddr)
	case LoginTypePhone:
		tokensFetchFunc = a.db.PhoneTokens
		updateLoginFunc = a.db.UpdateUserPhone
		usr, _, err = a.db.UserByPhone(forAddr)
	default:
		return nil, errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

	if err != nil {
		return nil, errors.Newf("get user by %s: %v", loginType, err)
	}

	tkn, err := a.dbTokenValid(usr.ID, dbt, tokensFetchFunc)
	if err != nil {
		return nil, err
	}
	phone, err := updateLoginFunc(usr.ID, tkn.Address, true)
	if err != nil {
		return nil, errors.Newf("update phone to verified: %v", err)
	}
	return phone, nil
}

func (a *Authentication) genAndSendTokens(tx *sql.Tx, action, loginType, toAddr, usrID string) (*DBTStatus, error) {

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
		return nil, errors.NewClientf(actionNotSupportedErrorF, action)
	}

	expiry := time.Now().Add(validity)

	URL := ""
	if a.webAppURLNilable != nil {
		tkn, err := a.genAndInsertToken(tx, expiry, loginType, usrID, toAddr)
		if err != nil {
			return nil, err
		}
		useURL := new(url.URL)
		*useURL = *a.webAppURLNilable // make copy so that a.webAppURLNilable remains pristine
		useURL.Path = path.Join(useURL.Path, action, loginType, toAddr, string(tkn))
		URL = useURL.String()
	}

	code, err := a.genAndInsertCode(tx, expiry, loginType, usrID, toAddr)
	if err != nil {
		return nil, err
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
		return nil, errors.Newf(actionNotSupportedErrorF, action)
	}

	var obfuscateFunc func(string) string
	tpl := a.loginTpActionTplts[loginType][action]

	switch loginType {
	case LoginTypePhone:
		obfuscateFunc = obfuscatePhone
		err = a.sendSMS(toAddr, tpl, sendData)
	case LoginTypeEmail:
		obfuscateFunc = obfuscateEmail
		err = a.sendEmail(toAddr, subj, tpl, sendData)
	default:
		return nil, errors.Newf(loginTypeNotSupportedErrorF, loginType)
	}
	if err != nil {
		return nil, err
	}

	return &DBTStatus{
		ObfuscatedAddress: obfuscateFunc(toAddr),
		ExpiresAt:         expiry,
	}, nil
}

func (a *Authentication) registerSelf(userType string, id string, password []byte, rcf regConditions, rf regFunc) (*User, error) {

	if !a.allowSelfReg {
		return nil, errors.NewForbidden("registration closed from the public")
	}

	if !inStrs(userType, validUserTypes) {
		return nil, errors.NewClientf("accountType must be one of %+v", validUserTypes)
	}

	passH, err := hashIfValid(password)
	if err != nil {
		return nil, err
	}

	id, err = rcf(id)
	if err != nil {
		return nil, err
	}

	grp, err := a.getOrCreateGroup(GroupPublic, AccessLevelPublic)
	if err != nil {
		return nil, err
	}
	ut, err := a.getOrCreateUserType(userType)
	if err != nil {
		return nil, err
	}

	usr := new(User)
	err = a.db.ExecuteTx(func(tx *sql.Tx) error {
		usr, err = a.db.InsertUserAtomic(tx, *ut, passH)
		if err != nil {
			return errors.Newf("insert user: %v", err)
		}
		if err := a.db.AddUserToGroupAtomic(tx, usr.ID, grp.ID); err != nil {
			return errors.Newf("add user to group: %v", err)
		}
		if err := rf(tx, ActionVerify, id, usr); err != nil {
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

func (a *Authentication) registerOther(regerLrgstGrp Group, userType, groupID, id string, pass []byte, rcf regConditions, f regFunc, actionType ...string) (*User, error) {

	if !inStrs(userType, validUserTypes) {
		return nil, errors.NewClientf("accountType must be one of %+v", validUserTypes)
	}
	if groupID == "" {
		return nil, errors.NewClientf("new user's group ID was not specified")
	}
	usrGroup, err := a.db.Group(groupID)
	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewClient("groupID does not exist")
		}
		return nil, errors.Newf("get group by ID: %v", err)
	}
	if regerLrgstGrp.AccessLevel > usrGroup.AccessLevel {
		return nil, errors.NewForbiddenf("You do not have enough rights to perform this operation")
	}
	id, err = rcf(id)
	if err != nil {
		return nil, err
	}

	ut, err := a.getOrCreateUserType(userType)
	if err != nil {
		return nil, err
	}

	passH, err := hashIfValid(pass)
	if err != nil {
		return nil, err
	}

	usr := new(User)
	err = a.db.ExecuteTx(func(tx *sql.Tx) error {
		usr, err = a.db.InsertUserAtomic(tx, *ut, passH)
		if err != nil {
			return errors.Newf("insert user: %v", err)
		}
		if err := a.db.AddUserToGroupAtomic(tx, usr.ID, usrGroup.ID); err != nil {
			return errors.Newf("add user to group: %v", err)
		}
		aT := ActionInvite
		if len(actionType) > 0 {
			aT = actionType[0]
		}
		if err := f(tx, aT, id, usr); err != nil {
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

func (a *Authentication) regUsernameConditions(username string) (string, error) {
	if username == "" {
		return "", errors.NewClient("username cannot be empty")
	}
	_, _, err := a.db.UserByUsername(username)
	return username, a.usrIdentifierAvail(LoginTypeUsername, err)
}

func (a *Authentication) regUsername(tx *sql.Tx, actionType, username string, usr *User) error {
	uname, err := a.db.InsertUserNameAtomic(tx, usr.ID, username)
	if err != nil {
		return errors.Newf("insert username: %v", err)
	}
	usr.UserName = *uname
	return nil
}

func (a *Authentication) regDevConditions(devID string) (string, error) {
	_, _, err := a.db.UserByDeviceID(devID)
	return devID, a.usrIdentifierAvail(LoginTypeDev, err)
}

func (a *Authentication) regDevice(tx *sql.Tx, actionType, devID string, usr *User) error {
	dev, err := a.db.InsertUserDeviceAtomic(tx, usr.ID, devID)
	if err != nil {
		return errors.Newf("insert device: %v", err)
	}
	usr.Devices = []Device{*dev}
	return nil
}

func (a *Authentication) regPhoneConditions(number string) (string, error) {
	number, err := formatValidPhone(number)
	if err != nil {
		return "", err
	}
	_, _, err = a.db.UserByPhone(number)
	return number, a.usrIdentifierAvail(LoginTypePhone, err)
}

func (a *Authentication) regPhone(tx *sql.Tx, actionType, number string, usr *User) error {
	phone, err := a.db.InsertUserPhoneAtomic(tx, usr.ID, number, false)
	if err != nil {
		return errors.Newf("insert phone: %v", err)
	}
	usr.Phone = *phone
	if a.smserNilable == nil {
		return nil
	}
	_, err = a.genAndSendTokens(tx, actionType, LoginTypePhone, number, usr.ID)
	return err
}

func (a *Authentication) regEmailConditions(email string) (string, error) {
	if email == "" {
		return "", errors.NewClient("email address cannot be empty")
	}
	_, _, err := a.db.UserByEmail(email)
	return email, a.usrIdentifierAvail(LoginTypeEmail, err)
}

func (a *Authentication) regEmail(tx *sql.Tx, actionType, address string, usr *User) error {
	email, err := a.db.InsertUserEmailAtomic(tx, usr.ID, address, false)
	if err != nil {
		return errors.Newf("insert email: %v", err)
	}
	usr.Email = *email
	if a.mailerNilable == nil {
		return nil
	}
	_, err = a.genAndSendTokens(tx, actionType, LoginTypeEmail, address, usr.ID)
	return err
}

func (a *Authentication) regFacebookConditions(fbToken string) (string, error) {
	if fbToken == "" {
		return "", errors.NewClient("facebook token cannot be empty")
	}
	fbID, err := a.validateFbToken(fbToken)
	if err != nil {
		return "", err
	}
	_, err = a.db.UserByFacebook(fbID)
	return fbID, a.usrIdentifierAvail(LoginTypeFacebook, err)
}

func (a *Authentication) regFacebook(tx *sql.Tx, actionType, fbID string, usr *User) error {
	fb, err := a.db.InsertUserFbIDAtomic(tx, usr.ID, fbID, true)
	if err != nil {
		return errors.Newf("insert facebook: %v", err)
	}
	usr.Facebook = *fb
	return nil
}

func (a *Authentication) updateUsername(usrID, newUsrName string) (*Username, error) {
	// TODO
	//_, _, err := a.db.UserByUsername(newUsrName)
	//if err = a.usrIdentifierAvail(LoginTypeUsername, err); err != nil {
	//	return nil, err
	//}
	//uname, err := a.db.UpdateUserName(clm.UsrID, newUsrName)
	//if err != nil {
	//	return nil, errors.Newf("user by phone: %v", err)
	//}
	return nil, errors.NewNotImplemented()
}

func (a *Authentication) updatePhone(usrID, newNum string) (*VerifLogin, error) {
	newNum, err := formatValidPhone(newNum)
	if err != nil {
		return nil, err
	}
	_, _, err = a.db.UserByPhone(newNum)
	if err = a.usrIdentifierAvail(LoginTypePhone, err); err != nil {
		return nil, err
	}
	phone, err := a.db.UpdateUserPhone(usrID, newNum, false)
	if err != nil {
		return nil, errors.Newf("update phone: %v", err)
	}
	if a.smserNilable == nil {
		return phone, nil
	}
	_, err = a.genAndSendTokens(nil, ActionVerify, LoginTypePhone, newNum, usrID)
	if err != nil {
		return nil, err
	}
	return phone, nil
}

func (a *Authentication) genAndInsertToken(tx *sql.Tx, expiry time.Time, loginType, forUsrID, loginID string) ([]byte, error) {
	dbt, err := a.urlTokenGen.SecureRandomBytes(56)
	if err != nil {
		return nil, errors.Newf("generate %s verification db token", loginType)
	}
	if err := a.hashAndInsertToken(tx, expiry, loginType, forUsrID, loginID, dbt); err != nil {
		return nil, errors.Newf("%s verification db token: %v", loginType, err)
	}
	return dbt, nil
}

func (a *Authentication) genAndInsertCode(tx *sql.Tx, expiry time.Time, loginType, forUsrID, loginID string) ([]byte, error) {
	code, err := a.numGen.SecureRandomBytes(6)
	if err != nil {
		return nil, errors.Newf("generate %s verification code", loginType)
	}
	if err := a.hashAndInsertToken(tx, expiry, loginType, forUsrID, loginID, code); err != nil {
		return nil, errors.Newf("%s verification code: %v", loginType, err)
	}
	return code, nil
}

func (a *Authentication) hashAndInsertToken(tx *sql.Tx, expiry time.Time, loginType, forUsrID, loginID string, code []byte) error {
	codeH, err := hash(code)
	if err != nil {
		return errors.Newf("hash for storage: %v", err)
	}

	var insFuncAtomic func(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)
	var insFunc func(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*DBToken, error)

	switch loginType {
	case LoginTypePhone:
		insFunc = a.db.InsertPhoneToken
		insFuncAtomic = a.db.InsertPhoneTokenAtomic
	case LoginTypeEmail:
		insFunc = a.db.InsertEmailToken
		insFuncAtomic = a.db.InsertEmailTokenAtomic
	default:
		return errors.NewClientf(loginTypeNotSupportedErrorF, loginType)
	}

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

func (a *Authentication) dbTokenValid(userID string, checkDBT []byte,
	f func(userID string, offset, count int64) ([]DBToken, error)) (*DBToken, error) {

	if len(checkDBT) == 0 {
		return nil, errors.NewUnauthorized("confirmation token cannot be empty")
	}
	if userID == "" {
		return nil, errors.NewUnauthorized("userID cannot be empty")
	}
	offset := int64(0)
	count := int64(100)
	var dbt DBToken
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
			err = bcrypt.CompareHashAndPassword(dbt.Token, checkDBT)
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
	return &dbt, nil
}

func (a *Authentication) getOrCreateGroup(groupName string, acl float32) (*Group, error) {
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

func (a *Authentication) validateJWTInOneOf(JWT string, gs ...Group) (JWTClaim, error) {
	clm := JWTClaim{}
	if JWT == "" {
		return clm, errors.NewUnauthorizedf("JWT must be provided")
	}
	_, err := a.jwter.Validate(JWT, &clm)
	if err != nil {
		return clm, errors.NewForbidden("invalid JWT")
	}
	for _, g := range gs {
		if inGroups(g, clm.Groups) {
			return clm, nil
		}
	}
	return clm, errors.NewForbidden("invalid JWT")
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

func hashIfValid(password []byte) ([]byte, error) {
	if len(password) < minPassLen {
		return nil, errors.NewClientf("password must be at least %d characters", minPassLen)
	}
	return hash(password)
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
