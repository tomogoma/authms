package model

import (
	"errors"
	"html/template"
	"reflect"

	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/generator"
)

// Option is used by NewAuthentication to pass additional configuration. Use
// the With... methods to create Options e.g.
//     nameOpt := WithAppName("My Awesome App")
type Option func(*authenticationConfig) error

// WithPasswordGen sets the password generator to be used. It cannot be nil;
func WithPasswordGen(g SecureRandomByteser, initErr error) Option {
	return func(c *authenticationConfig) error {
		if initErr != nil {
			return initErr
		}
		if g == nil || reflect.ValueOf(g).IsNil() {
			return errors.New("password Generator cannot be nil")
		}
		c.passGen = g
		return nil
	}
}

// WithNumGen sets the number generator to be used. It cannot be nil;
func WithNumGen(g SecureRandomByteser, intiErr error) Option {
	return func(c *authenticationConfig) error {
		if intiErr != nil {
			return intiErr
		}
		if g == nil || reflect.ValueOf(g).IsNil() {
			return errors.New("number Generator cannot be nil")
		}
		c.numGen = g
		return nil
	}
}

// WithURLTokenGen sets the string generator for URL tokens. It cannot be nil.
// Strings from this generator will be used in URLs and should thus conform to
// encoding rules.
func WithURLTokenGen(g SecureRandomByteser, intiErr error) Option {
	return func(c *authenticationConfig) error {
		if intiErr != nil {
			return intiErr
		}
		if g == nil || reflect.ValueOf(g).IsNil() {
			return errors.New("URL token Generator cannot be nil")
		}
		c.urlTokenGen = g
		return nil
	}
}

// WithDevLockedToUser requires a device ID during self-registration and only
// allows one user per device.
func WithDevLockedToUser(t bool) Option {
	return func(c *authenticationConfig) error {
		c.lockDevToUser = t
		return nil
	}
}

// WithSelfRegAllowed enables/disables self registration. Authenticate will fail
// if the value is set to false with no communication channel available.
// examples of communication channels are the Options WithSMSCl() or WithEmailCl()...
func WithSelfRegAllowed(t bool) Option {
	return func(c *authenticationConfig) error {
		c.allowSelfReg = t
		return nil
	}
}

// WithAppName sets the name of the application.
func WithAppName(n string) Option {
	return func(c *authenticationConfig) error {
		c.appNameEmptyable = n
		return nil
	}
}

// WithFacebookCl sets the facebook client to be used.
func WithFacebookCl(fb FacebookCl) Option {
	return func(c *authenticationConfig) error {
		c.fbNilable = fb
		return nil
	}
}

// WithSMSCl sets the SMS client to use.
// You may provide templates
// for sending SMSes e.g.
//     WithPhoneVerifyTplt()
// ...otherwise default template files in the config package are used.
// NewAuthentication() fails if this option is not nil and one of the
// templates couldn't be loaded.
func WithSMSCl(cl SMSer) Option {
	return func(c *authenticationConfig) error {
		c.smserNilable = cl
		return nil
	}
}

// WithEmailCl sets the email client to use.
// You may provide templates
// for sending Emails e.g.
//     WithEmailVerifyTplt()
// ...otherwise default template files in the config package are used.
// NewAuthentication() fails if this option is not nil and one of the
// templates couldn't be loaded.
func WithEmailCl(cl Mailer) Option {
	return func(c *authenticationConfig) error {
		c.mailerNilable = cl
		return nil
	}
}

// WithWebAppURL sets the webapp URL that will consume requests from users.
// TODO document types of request and URL endpoints that will be suffixed
func WithWebAppURL(URL string) Option {
	return func(c *authenticationConfig) error {
		c.webAppURLEmptyable = URL
		return nil
	}
}

// WithInvitationSubject sets the subject to be used when sending invites to a user.
func WithInvitationSubject(s string) Option {
	return func(c *authenticationConfig) error {
		c.invSubjEmptyable = s
		return nil
	}
}

// WithVerificationSubject sets the subject to be used when sending verification
// messages to a user.
func WithVerificationSubject(s string) Option {
	return func(c *authenticationConfig) error {
		c.verSubjEmptyable = s
		return nil
	}
}

// WithResetPassSubject sets the subject to be used when sending password reset
// codes/links to a user.
func WithResetPassSubject(s string) Option {
	return func(c *authenticationConfig) error {
		c.resPassSubjEmptyable = s
		return nil
	}
}

// WithPhoneInviteTplt sets the message template to be used when composing SMS
// invite messages.
// TODO define valid template values
func WithPhoneInviteTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil || reflect.ValueOf(t).IsNil() {
			return errors.New("provided phone invite template was nil")
		}
		c.loginTpActionTplts[loginTypePhone][ActionInvite] = t
		return nil
	}
}

// WithPhoneResetPassTplt sets the message template to be used when composing SMS
// password reset messages.
// TODO define valid template values
func WithPhoneResetPassTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil || reflect.ValueOf(t).IsNil() {
			return errors.New("provided phone invite template was nil")
		}
		c.loginTpActionTplts[loginTypePhone][ActionResetPass] = t
		return nil
	}
}

// WithPhoneVerifyTplt sets the message template to be used when composing phone
// verification messages.
// TODO define valid template values
func WithPhoneVerifyTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil || reflect.ValueOf(t).IsNil() {
			return errors.New("provided phone invite template was nil")
		}
		c.loginTpActionTplts[loginTypePhone][ActionVerify] = t
		return nil
	}
}

// WithEmailInviteTplt sets the message template to be used when composing email
// invite messages.
// TODO define valid template values
func WithEmailInviteTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil || reflect.ValueOf(t).IsNil() {
			return errors.New("provided phone invite template was nil")
		}
		c.loginTpActionTplts[loginTypeEmail][ActionInvite] = t
		return nil
	}
}

// WithEmailResetPassTplt sets the message template to be used when composing email
// password reset messages.
// TODO define valid template values
func WithEmailResetPassTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil || reflect.ValueOf(t).IsNil() {
			return errors.New("provided email reset pass template was nil")
		}
		c.loginTpActionTplts[loginTypeEmail][ActionResetPass] = t
		return nil
	}
}

// WithEmailVerifyTplt sets the message template to be used when composing email
// verification messages.
// TODO define valid template values
func WithEmailVerifyTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil || reflect.ValueOf(t).IsNil() {
			return errors.New("provided email verify template was nil")
		}
		c.loginTpActionTplts[loginTypeEmail][ActionVerify] = t
		return nil
	}
}

type authenticationConfig struct {
	// mandatory parameters
	passGen       SecureRandomByteser
	numGen        SecureRandomByteser
	urlTokenGen   SecureRandomByteser
	allowSelfReg  bool
	lockDevToUser bool
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

func (c *authenticationConfig) initializeValues() {
	c.allowSelfReg = true
	c.lockDevToUser = false
	c.loginTpActionTplts = map[string]map[string]*template.Template{
		loginTypePhone: make(map[string]*template.Template),
		loginTypeEmail: make(map[string]*template.Template),
	}
}

func (c *authenticationConfig) assignOptions(opts []Option) error {
	for _, optFunc := range opts {
		if err := optFunc(c); err != nil {
			return err
		}
	}
	return nil
}

func (c *authenticationConfig) fillDefaults() error {
	var defaultOpts []Option
	if c.passGen == nil {
		defaultOpts = append(
			defaultOpts,
			WithPasswordGen(generator.NewRandom(generator.AllChars)),
		)
	}
	if c.numGen == nil {
		defaultOpts = append(
			defaultOpts,
			WithPasswordGen(generator.NewRandom(generator.NumberChars)),
		)
	}
	if c.urlTokenGen == nil {
		defaultOpts = append(
			defaultOpts,
			WithPasswordGen(generator.NewRandom(generator.AlphaNumericChars)),
		)
	}
	if c.smserNilable != nil {
		phoneTpls := c.loginTpActionTplts[loginTypePhone]
		if _, ok := phoneTpls[ActionInvite]; !ok {
			defaultOpts = append(
				defaultOpts,
				WithPhoneInviteTplt(template.ParseFiles(config.DefaultPhoneInviteTpl)),
			)
		}
		if _, ok := phoneTpls[ActionVerify]; !ok {
			defaultOpts = append(
				defaultOpts,
				WithPhoneInviteTplt(template.ParseFiles(config.DefaultPhoneVerifyTpl)),
			)
		}
		if _, ok := phoneTpls[ActionResetPass]; !ok {
			defaultOpts = append(
				defaultOpts,
				WithPhoneInviteTplt(template.ParseFiles(config.DefaultPhoneResetPassTpl)),
			)
		}
	}
	if c.mailerNilable != nil {
		emailTPls := c.loginTpActionTplts[loginTypeEmail]
		if _, ok := emailTPls[ActionInvite]; !ok {
			defaultOpts = append(
				defaultOpts,
				WithPhoneInviteTplt(template.ParseFiles(config.DefaultEmailInviteTpl)),
			)
		}
		if _, ok := emailTPls[ActionVerify]; !ok {
			defaultOpts = append(
				defaultOpts,
				WithPhoneInviteTplt(template.ParseFiles(config.DefaultEmailVerifyTpl)),
			)
		}
		if _, ok := emailTPls[ActionResetPass]; !ok {
			defaultOpts = append(
				defaultOpts,
				WithPhoneInviteTplt(template.ParseFiles(config.DefaultEmailResetPassTpl)),
			)
		}
	}
	return c.assignOptions(defaultOpts)
}

func (c *authenticationConfig) valid() error {

	// with no means of communicating with the created user, the only means
	// of registration is self registration.
	// Communication is required in order to send tokens for setting up
	// account password and authentication.
	if c.mailerNilable == nil && c.smserNilable == nil && !c.allowSelfReg {
		return errors.New("preventing self registration with no means of" +
			" communication available. Nobody will be able to register")
	}
	return nil
}
