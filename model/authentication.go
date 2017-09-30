package model

import (
	"database/sql"
	"regexp"
	"strings"
	"time"

	"fmt"

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

	InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*Phone, error)
	InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*Email, error)
	InsertUserNameAtomic(tx *sql.Tx, userID, username string) (*Username, error)
	InsertUserFacebookIDAtomic(tx *sql.Tx, userID, fbID string, verified bool) (*Facebook, error)

	InsertPhoneToken(userID, phone string, token []byte, isUsed bool, expiry time.Time) (*PhoneToken, error)
	InsertEmailToken(userID, email string, token []byte, isUsed bool, expiry time.Time) (*EmailToken, error)
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

type APIKey struct {
	ID         string
	UserID     string
	APIKey     string
	CreateDate time.Time
	UpdateDate time.Time
}

type Authentication struct {
	allowSelfReg bool
	guard        Guard
	db           AuthStore
	fb           FacebookCl
	passGen      SecureRandomByteser
	numGen       SecureRandomByteser
	urlTokenGen  SecureRandomByteser
	smser        SMSer
	newAccSMSFmt string
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

	regTokenValidity = 2 * time.Hour
)

var rePhone = regexp.MustCompile(numExp)
var validUserTypes = []string{UserTypeIndividual, UserTypeCompany}

func (a *Authentication) RegisterByUsername(clientID, apiKey, userType, username, password string) (*User, error) {
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

func (a *Authentication) CreateByPhone(clientID, apiKey, token, userType, number, groupID string) (*User, error) {
	return a.createUser(clientID, apiKey, token, userType, groupID, func(tx *sql.Tx, usr *User) error {
		if err := a.insertPhoneAtomic(tx, usr, number); err != nil {
			return err
		}
		SMSCode, err := a.numGen.SecureRandomBytes(6)
		if err != nil {
			return errors.Newf("generate phone verification code")
		}
		SMSCodeH, err := bcrypt.GenerateFromPassword(SMSCode, bcrypt.DefaultCost)
		if err != nil {
			return errors.Newf("hash token for storage: %v", err)
		}
		expiry := time.Now().Add(regTokenValidity)
		_, err = a.db.InsertPhoneToken(usr.ID, number, SMSCodeH, false, expiry)
		if err != nil {
			return errors.Newf("insert SMS reg code: %v", err)
		}
		SMS := fmt.Sprintf(a.newAccSMSFmt, SMSCode)
		if err := a.smser.SMS(number, SMS); err != nil {
			return errors.Newf("send sms: %v", err)
		}
		return nil
	})
}

func (a *Authentication) RegisterByPhone(clientID, apiKey, userType, number, password string) (*User, error) {
	return a.registerUser(clientID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		return a.insertPhoneAtomic(tx, usr, number)
	})
}

func (a *Authentication) CreateByEmail(clientID, apiKey, token, userType, address, groupID string) (*User, error) {
	return a.createUser(clientID, apiKey, token, userType, groupID, func(tx *sql.Tx, usr *User) error {
		if err := a.insertEmailAtomic(tx, usr, address); err != nil {
			return err
		}
		emailCode, err := a.urlTokenGen.SecureRandomBytes(56)
		if err != nil {
			return errors.Newf("generate phone verification code")
		}
		emailCodeH, err := bcrypt.GenerateFromPassword(emailCode, bcrypt.DefaultCost)
		if err != nil {
			return errors.Newf("hash token for storage: %v", err)
		}
		expiry := time.Now().Add(regTokenValidity)
		_, err = a.db.InsertEmailToken(usr.ID, address, emailCodeH, false, expiry)
		if err != nil {
			return errors.Newf("insert email reg code: %v", err)
		}

	})
}

func (a *Authentication) RegisterByEmail(clientID, apiKey, userType, address, password string) (*User, error) {
	return a.registerUser(clientID, apiKey, userType, password, func(tx *sql.Tx, usr *User) error {
		return a.insertEmailAtomic(tx, usr, address)
	})
}

func (a *Authentication) RegisterPublicFacebook(clientID, apiKey, userType, fbID, fbToken string) (*User, error) {
	passwordB, err := a.passGen.SecureRandomBytes(genPassLen)
	if err != nil {
		return nil, errors.Newf("generate password: %v", err)
	}
	return a.registerUser(clientID, apiKey, userType, string(passwordB), func(tx *sql.Tx, usr *User) error {
		if fbID == "" {
			return errors.NewClient("facebook ID cannot be empty")
		}
		if fbToken == "" {
			return errors.NewClient("facebook token cannot be empty")
		}
		oa, err := a.fb.ValidateToken(fbToken)
		if err != nil {
			if a.fb.IsAuthError(err) {
				return errors.NewAuthf("facebook: %v", err)
			}
			return errors.Newf("validate facebook token: %v", err)
		}
		if oa.UserID() != fbID {
			return errors.NewAuth("facebook OAuth token does not belong to the user claimed")
		}
		fb, err := a.db.InsertUserFacebookIDAtomic(tx, usr.ID, fbID, true)
		if err != nil {
			return errors.Newf("insert facebook: %v", err)
		}
		usr.Facebook = *fb
		return nil
	})
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
	if len(password) < minPassLen {
		return nil, errors.NewClientf("password must be at least %d characters", minPassLen)
	}

	grp, err := a.getOrCreateGroup(GroupPublic, AccessLevelPublic)
	if err != nil {
		return nil, err
	}
	ut, err := a.getOrCreateUserType(userType)
	if err != nil {
		return nil, err
	}
	passH, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Newf("hash password: %v", err)
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

	password, err := a.passGen.SecureRandomBytes(genPassLen)
	if err != nil {
		return nil, errors.Newf("generate password: %v", err)
	}
	passH, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Newf("hash password: %v", err)
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
	_, err := a.tokenG.Validate(token, &clm)
	if err != nil {
		return clm, errors.NewForbidden("invalid token")
	}
	if !inGroups(g, clm.Groups) {
		return clm, errors.NewForbidden("invalid token")
	}
	return clm, nil
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
