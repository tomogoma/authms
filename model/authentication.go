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

	UpdateUserPhone(userID, phone string, verified bool) (*Phone, error)
	UpdateUserEmail(userID, email string, verified bool) (*Email, error)

	UpdatePasswordAtomic(tx *sql.Tx, userID string, password []byte) error
	UpdateUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*Phone, error)
	UpdateUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*Email, error)

	InsertUserPhone(userID, phone string, verified bool) (*Phone, error)
	InsertUserEmail(userID, email string, verified bool) (*Email, error)
	InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*Phone, error)
	InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*Email, error)
	InsertUserNameAtomic(tx *sql.Tx, userID, username string) (*Username, error)
	InsertUserFacebookIDAtomic(tx *sql.Tx, userID, fbID string, verified bool) (*Facebook, error)

	InsertPhoneToken(userID, phone string, token []byte, isUsed bool, expiry time.Time) (*Token, error)
	InsertEmailToken(userID, email string, token []byte, isUsed bool, expiry time.Time) (*Token, error)
	InsertPhoneTokenAtomic(tx *sql.Tx, userID, phone string, token []byte, isUsed bool, expiry time.Time) (*Token, error)
	InsertEmailTokenAtomic(tx *sql.Tx, userID, email string, token []byte, isUsed bool, expiry time.Time) (*Token, error)
	PhoneTokens(userID string, offset, count int64) ([]*Token, error)
	EmailTokens(userID string, offset, count int64) ([]*Token, error)

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

type Tokener interface {
	Generate(claims jwt.Claims) (string, error)
	Validate(token string, claims jwt.Claims) (*jwt.Token, error)
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
	tokener      Tokener
	mailer       Mailer

	invitationSubject    string
	verificationSubject  string
	emailInvitationTpl   *template.Template
	emailVerificationTpl *template.Template
	emailPassResetTpl    *template.Template
	phoneInvitationTpl   *template.Template
	phoneVerificationTpl *template.Template
	phonePassResetTpl    *template.Template
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

	numExp = `[0-9]+`

	minPassLen = 8
	genPassLen = 32

	inviteTokenValidity = 2 * time.Hour

	pathConfirm = "confirm"
	pathVerify  = "verify"

	loginTypeUsername = "username"
	loginTypeEmail    = "phone"
	loginTypePhone    = "email"
	loginTypeFacebook = "facebook"

	msgTypeVerification = "verification"
	msgTypePassReset    = "pass_reset"
)

var rePhone = regexp.MustCompile(numExp)
var validUserTypes = []string{UserTypeIndividual, UserTypeCompany}

func NewAuthentication() (*Authentication, error) {
	a := Authentication{}
	var err error
	a.emailInvitationTpl, err = template.ParseFiles(config.DefaultEmailInviteTpl)
	if err != nil {
		return nil, errors.Newf("read new account email template: %v", err)
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
		SMSURLToken, err := a.generateAndInsertToken(tx, loginTypePhone, usr.ID, number)
		if err != nil {
			return err
		}
		SMSURL, err := a.generateVerificationURL(loginTypePhone, usr.ID, SMSURLToken)
		if err != nil {
			return err
		}
		SMSCode, err := a.generateAndInsertCode(tx, loginTypePhone, usr.ID, number)
		if err != nil {
			return err
		}
		if err := a.sendPhoneVerification(a.phoneVerificationTpl, number, SMSURL, string(SMSCode)); err != nil {
			return err
		}
		return nil
	})
}

func (a *Authentication) RegisterPublicByEmail(clientID, apiKey, userType, address, password string) (*User, error) {
	return a.registerUser(clientID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		if err := a.insertEmailAtomic(tx, usr, address); err != nil {
			return err
		}
		token, err := a.generateAndInsertToken(tx, loginTypeEmail, usr.ID, address)
		if err != nil {
			return err
		}
		verURL, err := a.generateVerificationURL(loginTypeEmail, usr.ID, token)
		if err != nil {
			return err
		}
		code, err := a.generateAndInsertCode(tx, loginTypeEmail, usr.ID, address)
		if err != nil {
			return err
		}
		if err := a.sendEmailVerification(a.emailVerificationTpl, address, verURL, string(code)); err != nil {
			return err
		}
		return nil
	})
}

func (a *Authentication) RegisterPublicByFacebook(clientID, apiKey, userType, fbToken string) (*User, error) {
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

func (a *Authentication) CreateByPhone(clientID, apiKey, token, userType, number, groupID string) (*User, error) {
	if a.smser == nil {
		return nil, errors.NewNotImplementedf("SMS notification (to created user) not available")
	}
	return a.createUser(clientID, apiKey, token, userType, groupID, func(tx *sql.Tx, usr *User) error {
		if err := a.insertPhoneAtomic(tx, usr, number); err != nil {
			return err
		}
		SMSURLToken, err := a.generateAndInsertToken(tx, loginTypePhone, usr.ID, number)
		if err != nil {
			return err
		}
		SMSURL, err := a.generateConfirmationURL(loginTypePhone, usr.ID, SMSURLToken)
		if err != nil {
			return err
		}
		if err := a.sendPhoneInvite(number, SMSURL); err != nil {
			return err
		}
		return nil
	})
}

func (a *Authentication) CreateByEmail(clientID, apiKey, token, userType, address, groupID string) (*User, error) {
	if a.mailer == nil {
		return nil, errors.NewNotImplementedf("email notification (to created user) not available")
	}
	return a.createUser(clientID, apiKey, token, userType, groupID, func(tx *sql.Tx, usr *User) error {
		if err := a.insertEmailAtomic(tx, usr, address); err != nil {
			return err
		}
		token, err := a.generateAndInsertToken(tx, loginTypeEmail, usr.ID, address)
		if err != nil {
			return err
		}
		verURL, err := a.generateConfirmationURL(loginTypeEmail, usr.ID, token)
		if err != nil {
			return err
		}
		if err := a.sendEmailInvite(address, verURL); err != nil {
			return err
		}
		return nil
	})
}

func (a *Authentication) ConfirmPhoneAccount(clientID, apiKey, userID, token, password string) (*Phone, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	tkn, err := a.phoneTokenValid(userID, token)
	if err != nil {
		return nil, err
	}
	passH, err := hashIfValid(password)
	if err != nil {
		return nil, err
	}
	var phone *Phone
	err = a.db.ExecuteTx(func(tx *sql.Tx) error {
		err = a.db.UpdatePasswordAtomic(tx, userID, passH)
		if err != nil {
			return errors.Newf("update password: %v", err)
		}
		phone, err = a.db.UpdateUserPhoneAtomic(tx, userID, tkn.Phone, true)
		if err != nil {
			return errors.Newf("update phone to verified: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return phone, nil
}

func (a *Authentication) ConfirmEmailAccount(clientID, apiKey, userID, token, password string) (*Email, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	tkn, err := a.emailTokenValid(userID, token)
	if err != nil {
		return nil, err
	}
	passH, err := hashIfValid(password)
	if err != nil {
		return nil, err
	}
	var email *Email
	err = a.db.ExecuteTx(func(tx *sql.Tx) error {
		err = a.db.UpdatePasswordAtomic(tx, userID, passH)
		if err != nil {
			return errors.Newf("update password: %v", err)
		}
		email, err = a.db.UpdateUserEmailAtomic(tx, userID, tkn.Email, true)
		if err != nil {
			return errors.Newf("update phone to verified: %v", err)
		}
		return nil
	})
	return email, nil
}

func (a *Authentication) SendMsg(clientID, apiKey, messageType, loginType, jwt, identifier string) error {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return err
	}
	clms := new(Claim)
	if _, err := a.tokener.Validate(jwt, clms); err != nil {
		return errors.NewForbiddenf("you have no access")
	}

	var err error
	var usr *User
	switch loginType {
	case loginTypePhone:
		usr, _, err = a.db.UserByPhone(identifier)
	case loginTypeEmail:
		usr, _, err = a.db.UserByEmail(identifier)
	default:
		return errors.NewClientf("login type not supported: %v", loginType)
	}

	if err != nil {
		if !a.db.IsNotFoundError(err) {
			return errors.Newf("user by identifier: %v", err)
		}
		usr, _, err = a.db.User(clms.UsrID)
		if err != nil {
			if a.db.IsNotFoundError(err) {
				return errors.Newf("valid claims with non-exist user found: %+v", clms)
			}
			return errors.Newf("get user: %v", err)
		}
		switch loginType {
		case loginTypePhone:
			_, err = a.db.InsertUserPhone(usr.ID, identifier, false)
		case loginTypeEmail:
			_, err = a.db.InsertUserEmail(usr.ID, identifier, false)
		default:
			return errors.Newf("login type not supported: %v", loginType)
		}
		if err != nil {
			return errors.Newf("insert identifier: %v", err)
		}
	}
	tkn, err := a.generateAndInsertToken(nil, loginTypePhone, usr.ID, identifier)
	if err != nil {
		return err
	}
	SMSURL, err := a.generateConfirmationURL(loginTypePhone, usr.ID, tkn)
	if err != nil {
		return err
	}
	code, err := a.generateAndInsertCode(nil, loginTypePhone, usr.ID, identifier)
	if err != nil {
		return err
	}
	var phoneTpl *template.Template
	var emailTpl *template.Template
	switch messageType {
	case msgTypePassReset:
		phoneTpl = a.phonePassResetTpl
		emailTpl = a.emailPassResetTpl
	case msgTypeVerification:
		phoneTpl = a.phoneVerificationTpl
		emailTpl = a.emailVerificationTpl
	default:
		return errors.Newf("message type unsupported: %v", messageType)
	}
	switch loginType {
	case loginTypePhone:
		err = a.sendPhoneVerification(phoneTpl, identifier, SMSURL, string(code))
	case loginTypeEmail:
		err = a.sendEmailVerification(emailTpl, identifier, SMSURL, string(code))
	default:
		return errors.Newf("login type not supported: %v", loginType)
	}
	if err != nil {
		return err
	}
	return nil
}

func (a *Authentication) VerifyPhone(clientID, apiKey, userID, token string) (*Phone, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	tkn, err := a.phoneTokenValid(userID, token)
	if err != nil {
		return nil, err
	}
	phone, err := a.db.UpdateUserPhone(userID, tkn.Phone, true)
	if err != nil {
		return nil, errors.Newf("update phone to verified: %v", err)
	}
	return phone, nil
}

func (a *Authentication) VerifyEmail(clientID, apiKey, userID, token string) (*Email, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	tkn, err := a.emailTokenValid(userID, token)
	if err != nil {
		return nil, err
	}
	email, err := a.db.UpdateUserEmail(userID, tkn.Email, true)
	if err != nil {
		return nil, errors.Newf("update phone to verified: %v", err)
	}
	return email, nil
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
	default:
		return nil, errors.Newf("unknown login type %s", loginType)
	}
	if err != nil {
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewUnauthorizedf("invalid username/password combination")
		}
		return nil, errors.Newf("get user by %s: %v", loginType, err)
	}
	if err := passwordValid(passHB, []byte(password)); err != nil {
		return nil, errors.NewUnauthorized("invalid username/password combination")
	}
	return usr, nil
}

func (a *Authentication) LoginFacebook(clientID, apiKey, fbToken string) (*User, error) {
	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
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

func (a *Authentication) SendPasswordResetToken(c context.Context, loginType, withAddress string) (string, error) {
	return "", errors.NewNotImplemented()
}

func (a *Authentication) ResetPassword(c context.Context, loginType, resetToken, newPassword string) error {
	return errors.NewNotImplemented()
}

func (a *Authentication) UpdatePhoneLogin(c context.Context, userID, phone string) (*User, error) {
	return nil, errors.NewNotImplemented()
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

func (a *Authentication) createUser(clientID, apiKey, token, userType,
	groupID string, f func(tx *sql.Tx, usr *User) error) (*User, error) {

	if err := a.guard.APIKeyValid(clientID, apiKey); err != nil {
		return nil, err
	}
	adminGrp, err := a.getOrCreateGroup(GroupAdmin, AccessLevelAdmin)
	if err != nil {
		return nil, err
	}
	if _, err := a.validateTokenInGroup(token, *adminGrp); err != nil {
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

func (a *Authentication) generateConfirmationURL(loginType, userID string, token []byte) (string, error) {
	verURL, err := url.Parse(a.webAppURL)
	if err != nil {
		return "", errors.Newf("parsing web app URL: %v", err)
	}
	verURL.Path = path.Join(verURL.Path, pathConfirm, loginType, userID, string(token))
	return verURL.String(), nil
}

func (a *Authentication) generateVerificationURL(loginType, userID string, token []byte) (string, error) {
	verURL, err := url.Parse(a.webAppURL)
	if err != nil {
		return "", errors.Newf("parsing web app URL: %v", err)
	}
	verURL.Path = path.Join(verURL.Path, pathVerify, loginType, userID, string(token))
	return verURL.String(), nil
}

func (a *Authentication) sendEmailInvite(toAddress, verURL string) error {
	emailBf := bytes.NewBuffer(make([]byte, 0, 256))
	err := a.emailInvitationTpl.Execute(emailBf, InvitationTemplate{
		AppName:  a.appName,
		URLToken: verURL,
	})
	if err != nil {
		return errors.Newf("email body from template: %v", err)
	}
	err = a.mailer.SendEmail(SendMail{
		ToEmails: []string{toAddress},
		Subject:  a.invitationSubject,
		Body:     template.HTML(emailBf.String()),
	})
	if err != nil {
		return errors.Newf("send email: %v", err)
	}
	return nil
}

func (a *Authentication) sendEmailVerification(t *template.Template, toAddress, verURL, verCode string) error {
	emailBf := bytes.NewBuffer(make([]byte, 0, 256))
	err := t.Execute(emailBf, VerificationTemplate{
		AppName:  a.appName,
		URLToken: verURL,
		Code:     verCode,
	})
	if err != nil {
		return errors.Newf("email body from template: %v", err)
	}
	err = a.mailer.SendEmail(SendMail{
		ToEmails: []string{toAddress},
		Subject:  a.invitationSubject,
		Body:     template.HTML(emailBf.String()),
	})
	if err != nil {
		return errors.Newf("send email: %v", err)
	}
	return nil
}

func (a *Authentication) generateAndInsertToken(tx *sql.Tx, loginType, forUsrID, loginID string) ([]byte, error) {
	token, err := a.passGen.SecureRandomBytes(56)
	if err != nil {
		return nil, errors.Newf("generate %s verification token", loginType)
	}
	if err := a.hashAndInsertToken(tx, loginType, forUsrID, loginID, token); err != nil {
		return nil, errors.Newf("%s verification token: %v", loginType, err)
	}
	return token, nil
}

func (a *Authentication) generateAndInsertCode(tx *sql.Tx, loginType, forUsrID, loginID string) ([]byte, error) {
	code, err := a.numGen.SecureRandomBytes(6)
	if err != nil {
		return nil, errors.Newf("generate %s verification code", loginType)
	}
	if err := a.hashAndInsertToken(tx, loginType, forUsrID, loginID, code); err != nil {
		return nil, errors.Newf("%s verification code: %v", loginType, err)
	}
	return code, nil
}

func (a *Authentication) hashAndInsertToken(tx *sql.Tx, loginType, forUsrID, loginID string, code []byte) error {
	codeH, err := hash([]byte(code))
	if err != nil {
		return errors.Newf("hash for storage: %v", err)
	}
	expiry := time.Now().Add(inviteTokenValidity)
	if tx == nil {
		switch loginType {
		case loginTypePhone:
			_, err = a.db.InsertPhoneToken(forUsrID, loginID, codeH, false, expiry)
		case loginTypeEmail:
			_, err = a.db.InsertEmailToken(forUsrID, loginID, codeH, false, expiry)
		default:
			return errors.Newf("unsupported loginType")
		}
	} else {
		switch loginType {
		case loginTypePhone:
			_, err = a.db.InsertPhoneTokenAtomic(tx, forUsrID, loginID, codeH, false, expiry)
		case loginTypeEmail:
			_, err = a.db.InsertEmailTokenAtomic(tx, forUsrID, loginID, codeH, false, expiry)
		default:
			return errors.Newf("unsupported loginType")
		}
	}
	if err != nil {
		return errors.Newf("insert: %v", err)
	}
	return nil
}

func (a *Authentication) sendPhoneInvite(toPhone, urlToken string) error {
	SMSBf := bytes.NewBuffer(make([]byte, 0, 256))
	err := a.phoneInvitationTpl.Execute(SMSBf, InvitationTemplate{
		AppName:  a.appName,
		URLToken: urlToken,
	})
	if err != nil {
		return errors.Newf("SMS from template: %v", err)
	}
	if err := a.smser.SMS(toPhone, SMSBf.String()); err != nil {
		return errors.Newf("send SMS: %v", err)
	}
	return nil
}

func (a *Authentication) sendPhoneVerification(t *template.Template, toPhone, urlToken, smsCode string) error {
	SMSBf := bytes.NewBuffer(make([]byte, 0, 256))
	err := t.Execute(SMSBf, VerificationTemplate{
		AppName:  a.appName,
		URLToken: urlToken,
		Code:     smsCode,
	})
	if err != nil {
		return errors.Newf("SMS from template: %v", err)
	}
	if err := a.smser.SMS(toPhone, SMSBf.String()); err != nil {
		return errors.Newf("send SMS: %v", err)
	}
	return nil
}

func (a *Authentication) phoneTokenValid(userID, token string) (*Token, error) {
	return a.tokenValid(userID, token, a.db.PhoneTokens)
}

func (a *Authentication) emailTokenValid(userID, token string) (*Token, error) {
	return a.tokenValid(userID, token, a.db.EmailTokens)
}

func (a *Authentication) tokenValid(userID, token string,
	f func(userID string, offset, count int64) ([]*Token, error)) (*Token, error) {
	if token == "" {
		return nil, errors.NewUnauthorized("confirmation token cannot be empty")
	}
	if userID == "" {
		return nil, errors.NewUnauthorized("userID cannot be empty")
	}
	offset := int64(0)
	count := int64(100)
	var tkn *Token
resumeFunc:
	for {
		codes, err := f(userID, offset, count)
		if a.db.IsNotFoundError(err) {
			return nil, errors.NewForbidden("token is invalid")
		}
		if err != nil {
			return nil, errors.Newf("get phone tokens: %v", err)
		}
		offset = offset + count
		for _, tkn = range codes {
			err = bcrypt.CompareHashAndPassword(tkn.Token, []byte(token))
			if err != nil {
				continue
			}
			break resumeFunc
		}
	}
	if tkn.IsUsed {
		return nil, errors.NewForbiddenf("token already used")
	}
	if time.Now().After(tkn.ExpiryDate) {
		return nil, errors.NewAuth("token has expired")
	}
	return tkn, nil
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

func (a *Authentication) validateTokenInGroup(token string, g Group) (Claim, error) {
	clm := Claim{}
	if token == "" {
		return clm, errors.NewUnauthorizedf("token must be provided")
	}
	_, err := a.tokener.Validate(token, &clm)
	if err != nil {
		return clm, errors.NewForbidden("invalid token")
	}
	if !inGroups(g, clm.Groups) {
		return clm, errors.NewForbidden("invalid token")
	}
	return clm, nil
}

func passwordValid(hashed, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashed, password)
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
