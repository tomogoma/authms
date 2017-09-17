package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/generator"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/errors"
	"golang.org/x/crypto/bcrypt"
)

type OAuthHandler interface {
	IsAuthError(error) bool
	ValidateToken(appName, token string) (OAuthResponse, error)
}

type TokenGenerator interface {
	Generate(claims jwt.Claims) (string, error)
	Validate(token string, claims jwt.Claims) (*jwt.Token, error)
}

type OAuthClient interface {
	IsAuthError(error) bool
	ValidateToken(string) (OAuthResponse, error)
}

type SecureRandomByteser interface {
	SecureRandomBytes(length int) ([]byte, error)
}

type DBHelper interface {
	UserExists(u *authms.User) (int64, error)
	SaveUser(*authms.User) error
	SaveHistory(*authms.History) error
	GetByUserName(uname, pass string) (*authms.User, error)
	GetByAppUserID(appName, appUserID string) (*authms.User, error)
	GetByPhone(phone, pass string) (*authms.User, error)
	GetByEmail(email, pass string) (*authms.User, error)
	GetHistory(userID int64, offset, count int, accessType ...string) ([]*authms.History, error)
	UpdatePhone(userID int64, newPhone *authms.Value) error
	UpdateAppUserID(userID int64, new *authms.OAuth) error
	UpsertLoginVerification(code LoginVerification) error
	GetLoginVerifications(verificationType string, userID, offset, count int64) ([]LoginVerification, error)
	IsNotFoundError(err error) bool
	IsDuplicateError(err error) bool
}

type SMSer interface {
	IsNotImplementedError(error) bool
	SMS(toPhone, message string) error
}

type LoginVerification struct {
	Type         string
	SubjectValue string
	UserID       int64
	Issue        time.Time
	Expiry       time.Time
	CodeHash     []byte
	IsUsed       bool
}

type Auth struct {
	dbHelper         DBHelper
	tokenG           TokenGenerator
	smser            SMSer
	randomSMSCodeser SecureRandomByteser
	oAuthCls         map[string]OAuthClient
	smsFmt           string
	smsValidity      time.Duration
	errors.ClErrCheck
	errors.AuthErrCheck
	errors.NotImplErrCheck
}

type Option func(*Auth)

const (
	numPrevLogins        = 5
	AccessLogin          = "LOGIN"
	AccessRegistration   = "REGISTER"
	AccessUpdate         = "UPDATE"
	AccessVerification   = "VERIFICATION"
	AccessCodeValidation = "VERIFICATION_CODE_VALIDATION"
	numExp               = `[0-9]+`
	tokenValidity        = 8 * time.Hour
	AppFacebook          = "facebook"

	verTypePhone = "phone"
)

var rePhone = regexp.MustCompile(numExp)

func WithFB(fb OAuthClient) Option {
	return func(a *Auth) {
		a.oAuthCls[AppFacebook] = fb
	}
}

func WithSMSFormat(fmt string) Option {
	return func(a *Auth) {
		a.smsFmt = fmt
	}
}

func WithSMSValidity(d time.Duration) Option {
	return func(a *Auth) {
		a.smsValidity = d
	}
}

func WithSMSCodeGeneratory(g SecureRandomByteser) Option {
	return func(a *Auth) {
		a.randomSMSCodeser = g
	}
}

func New(tg TokenGenerator, db DBHelper, sms SMSer, opts ...Option) (*Auth, error) {
	randSMSCodeser, err := generator.NewRandom(generator.NumberChars)
	if err != nil {
		return nil, errors.Newf("instantiate random number code generator: %v", err)
	}
	SMSFmt := "Your verification code is %s"
	SMSValidity := 5 * time.Minute
	a := &Auth{
		dbHelper:         db,
		tokenG:           tg,
		smser:            sms,
		randomSMSCodeser: randSMSCodeser,
		oAuthCls:         make(map[string]OAuthClient),
		smsFmt:           SMSFmt,
		smsValidity:      SMSValidity,
	}
	for _, f := range opts {
		f(a)
	}
	if a.tokenG == nil {
		return nil, errors.New("token generator was nil")
	}
	if a.dbHelper == nil {
		return nil, errors.New("DBHelper was nil")
	}
	if a.smser == nil {
		return nil, errors.New("SMSer was nil")
	}
	if a.randomSMSCodeser == nil {
		return nil, errors.New("Random SMS code generator was nil")
	}
	if err = testMessageFormat(a.smsFmt); err != nil {
		return nil, err
	}
	if a.smsValidity < 1*time.Minute {
		return nil, errors.New("invalid sms validity duration: must be > 1m")
	}
	for name, OAuthCl := range a.oAuthCls {
		if OAuthCl == nil {
			return nil, errors.Newf("nil OAuth client (%s)", name)
		}
	}
	return a, nil
}

func (a *Auth) Register(user *authms.User, devID, rIP string) error {
	defer func() {
		if user != nil {
			user.Password = ""
		}
	}()
	if err := validateUser(user); err != nil {
		return err
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
		user.OAuths[appName].Verified = true
	}
	if user.Phone != nil {
		user.Phone.Value = formatPhone(user.Phone.Value)
		user.Phone.Verified = false
	}
	if user.Email != nil {
		user.Email.Verified = false
	}
	if err := a.checkUserExists(user); err != nil {
		return err
	}
	err := a.dbHelper.SaveUser(user)
	if err != nil {
		return errors.Newf("error persisting user: %v", err)
	}
	go a.saveHistory(user, devID, AccessRegistration, rIP, nil)
	return nil
}

func (a *Auth) UpdatePhone(user *authms.User, token, devID, rIP string) error {
	if user == nil {
		return errors.NewClient("user was empty")
	}
	var err error
	defer func() {
		go a.saveHistory(user, devID, AccessUpdate, rIP, err)
	}()
	if _, err = a.validateToken(user.ID, token); err != nil {
		return err
	}
	if !hasValue(user.Phone) {
		return errors.NewClient("phone was invalid")
	}
	if devID == "" {
		return errors.NewClient("device ID was empty")
	}
	user.Phone.Value = formatPhone(user.Phone.Value)
	user.Phone.Verified = false
	if err := a.checkUserExists(user); err != nil {
		return err
	}
	err = a.dbHelper.UpdatePhone(user.ID, user.Phone)
	if err != nil {
		return errors.Newf("error persisting phone update: %v", err)
	}
	return nil
}

func (a *Auth) VerifyPhone(req *authms.SMSVerificationRequest, rIP string) (*authms.SMSVerificationStatus, error) {
	if req == nil {
		return nil, errors.NewClient("SMSVerificationRequest was empty")
	}
	var err error
	defer func() {
		go a.saveHistory(&authms.User{ID: req.UserID}, req.DeviceID,
			AccessVerification, rIP, err)
	}()
	if _, err = a.validateToken(req.UserID, req.Token); err != nil {
		return nil, err
	}
	req.Phone = formatPhone(req.Phone)
	err = a.checkUserExists(&authms.User{
		ID:    req.UserID,
		Phone: &authms.Value{Value: req.Phone},
	})
	if err != nil {
		return nil, err
	}
	SMSCode, err := a.randomSMSCodeser.SecureRandomBytes(4)
	if err != nil {
		return nil, errors.Newf("generate SMS code: %v", err)
	}
	SMSCodeHash, err := bcrypt.GenerateFromPassword(SMSCode, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Newf("hash SMS code: %v", err)
	}
	issue := time.Now()
	expiry := issue.Add(a.smsValidity)
	err = a.dbHelper.UpsertLoginVerification(LoginVerification{
		Type:         verTypePhone,
		SubjectValue: req.Phone,
		Issue:        issue,
		Expiry:       expiry,
		UserID:       req.UserID,
		CodeHash:     SMSCodeHash,
		IsUsed:       false,
	})
	if err != nil {
		return nil, errors.Newf("persist SMS code: %v", err)
	}
	smsBody := fmt.Sprintf(a.smsFmt, SMSCode)
	if err = a.smser.SMS(req.Phone, smsBody); err != nil {
		if a.smser.IsNotImplementedError(err) {
			return nil, errors.NewNotImplementedf("%v", err)
		}
		return nil, err
	}
	return &authms.SMSVerificationStatus{
		Token:     req.Token,
		Phone:     req.Phone,
		ExpiresAt: expiry.Format(config.TimeFormat),
		Verified:  false,
	}, nil
}

func (a *Auth) VerifyPhoneCode(req *authms.SMSVerificationCodeRequest, rIP string) (*authms.SMSVerificationStatus, error) {
	if req == nil {
		return nil, errors.NewClient("SMSVerificationRequest was empty")
	}
	var err error
	defer func() {
		go a.saveHistory(&authms.User{ID: req.UserID}, req.DeviceID,
			AccessCodeValidation, rIP, err)
	}()
	if _, err = a.validateToken(req.UserID, req.Token); err != nil {
		return nil, err
	}
	offset := int64(0)
	count := int64(100)
	var verification LoginVerification
resumeFunc:
	for {
		codes, err := a.dbHelper.GetLoginVerifications(verTypePhone, req.UserID, offset, count)
		if a.dbHelper.IsNotFoundError(err) {
			return nil, errors.NewAuth("Invalid SMS code")
		}
		if err != nil {
			return nil, errors.Newf("get SMS codes: %v", err)
		}
		offset = offset + count
		for _, verification = range codes {
			err = bcrypt.CompareHashAndPassword(verification.CodeHash, []byte(req.Code))
			if err != nil {
				continue
			}
			break resumeFunc
		}
	}
	if verification.IsUsed {
		return nil, errors.NewAuth("code already used")
	}
	if time.Now().After(verification.Expiry) {
		return nil, errors.NewAuth("%code is expired")
	}
	err = a.dbHelper.UpdatePhone(
		req.UserID, &authms.Value{
			Verified: true,
			Value:    verification.SubjectValue,
		})
	if err != nil {
		return nil, errors.Newf("persist phone update: %v", err)
	}
	verification.IsUsed = true
	if err = a.dbHelper.UpsertLoginVerification(verification); err != nil {
		return nil, errors.Newf("upsert login verification: %v", err)
	}
	return &authms.SMSVerificationStatus{
		Phone:    verification.SubjectValue,
		Verified: true,
	}, nil
}

func (a *Auth) UpdateOAuth(user *authms.User, appName, token, devID, rIP string) error {
	if user == nil {
		return errors.NewClient("user was empty")
	}
	var err error
	defer func() {
		go a.saveHistory(user, devID, AccessUpdate, rIP, err)
	}()
	if _, err = a.validateToken(user.ID, token); err != nil {
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
	if err := a.checkUserExists(user); err != nil {
		return err
	}
	err = a.dbHelper.UpdateAppUserID(user.ID, oa)
	if err != nil {
		return errors.Newf("error persisting OAuth changes: %v", err)
	}
	return nil
}

func (a *Auth) UpdatePassword(userID int64, oldPass, newPass, devID, rIP string) error {
	return errors.NewNotImplemented()
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

func (a *Auth) LoginPhone(phone, pass, devID, rIP string) (*authms.User, error) {
	if devID == "" {
		return nil, errors.NewClient("Dev ID was empty")
	}
	phone = formatPhone(phone)
	usr, err := a.dbHelper.GetByPhone(phone, pass)
	if err = a.processLoginResults(usr, devID, rIP, err); err != nil {
		return nil, err
	}
	return usr, nil
}

func (a *Auth) LoginEmail(email, pass, devID, rIP string) (*authms.User, error) {
	if devID == "" {
		return nil, errors.NewClient("Dev ID was empty")
	}
	usr, err := a.dbHelper.GetByEmail(email, pass)
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

func (a *Auth) checkUserExists(user *authms.User) error {
	existUsrID, err := a.dbHelper.UserExists(user)
	if err != nil {
		return errors.Newf("error checking if user exists: %v", err)
	}
	if existUsrID >= 1 && user.ID != existUsrID {
		return errors.NewClient("A user with some of the provided" +
			" credentials already exists")
	}
	return nil
}

func (a *Auth) validateToken(userID int64, token string) (Claim, error) {
	clm := Claim{}
	_, err := a.tokenG.Validate(token, &clm)
	if err != nil {
		return clm, errors.NewAuthf("invalid token: %v", err)
	}
	if clm.UsrID != userID {
		return clm, errors.NewAuthf("token-user (userid) mismatch:"+
			" userID is %d token belongs to %d", userID, clm.UsrID)
	}
	return clm, nil
}

func (a *Auth) processLoginResults(usr *authms.User, devID, rIP string, loginErr error) error {
	if a.dbHelper.IsNotFoundError(loginErr) {
		loginErr = errors.NewAuth("invalid credentials")
	}
	defer func() {
		go a.saveHistory(usr, devID, AccessLogin, rIP, loginErr)
	}()
	if loginErr != nil {
		return loginErr
	}
	clm := NewClaim(usr.ID, devID, tokenValidity)
	tkn, loginErr := a.tokenG.Generate(clm)
	if loginErr != nil {
		loginErr = errors.Newf("error generating token: %v", loginErr)
		return loginErr
	}
	usr.Token = tkn
	prevLogins, loginErr := a.dbHelper.GetHistory(usr.ID, 0, numPrevLogins,
		AccessLogin)
	if loginErr != nil {
		loginErr = errors.Newf("error fetching login history: %v", loginErr)
		return loginErr
	}
	usr.LoginHistory = prevLogins
	return nil
}

func (a *Auth) validateOAuth(claimOA *authms.OAuth) error {
	if claimOA == nil {
		return errors.NewClient("nil OAuth was found")
	}
	cl, exists := a.oAuthCls[claimOA.AppName]
	if !exists {
		return errors.NewClient("the app provided is not supported")
	}
	oa, err := cl.ValidateToken(claimOA.AppToken)
	if err != nil {
		if cl.IsAuthError(err) {
			return errors.NewAuthf("%s: %v", claimOA.AppName, err)
		}
		return errors.Newf("validate token: %s: %v", claimOA.AppName, err)
	}
	if oa.UserID() != claimOA.AppUserID {
		return errors.NewAuth("OAuth token does not belong to the user claimed")
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
		logrus.WithFields(logrus.Fields{
			logging.FieldAction:  "save history entry",
			logging.FieldHistory: h,
		}).Error(err)
	}
}

func testMessageFormat(msgFmt string) error {
	r, err := regexp.Compile("%s")
	if err != nil {
		return errors.Newf("error compiling message tester regex: %v", err)
	}
	formatters := r.FindAllString(msgFmt, -1)
	if len(formatters) != 1 {
		return errors.Newf("Expected 1 '%%s' formatter but got %d", len(formatters))
	}
	return nil
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

func havePasswordComboAuth(u *authms.User) bool {
	return hasValue(u.Phone) || hasValue(u.Email) || u.UserName != ""
}

func validateUser(u *authms.User) error {
	if u == nil {
		return errors.NewClient("user was not provided")
	}
	hasPhone := hasValue(u.Phone)
	hasMail := hasValue(u.Email)
	if u.UserName == "" && !hasPhone && !hasMail && u.OAuths == nil {
		return errors.NewClient("A user must have at least one" +
			" identifier (UserName, Phone, Email, OAuthApp")
	}
	if havePasswordComboAuth(u) && u.Password == "" {
		return errors.NewClient("password was not provided")
	}
	for name, oa := range u.OAuths {
		if name != oa.AppName {
			return errors.NewClient("an OAuth key did not match AppName")
		}
		if err := validateOAuth(oa); err != nil {
			return err
		}
	}
	return nil
}

func validateOAuth(oa *authms.OAuth) error {
	if oa == nil {
		return errors.NewClient("OAuth was not provided")
	}
	if oa.AppName == "" {
		return errors.NewClient("AppName was not provided")
	}
	if oa.AppToken == "" {
		return errors.NewClient("AppToken was not provided")
	}
	if oa.AppUserID == "" {
		return errors.NewClient("AppUserID was not provided")
	}
	return nil
}

func hasValue(v *authms.Value) bool {
	return v != nil && v.Value != ""
}
