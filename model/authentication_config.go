package model

import (
	"errors"
	"html/template"

	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/generator"
)

type Option func(*authenticationConfig) error

func WithPasswordGen(g SecureRandomByteser, initErr error) Option {
	return func(c *authenticationConfig) error {
		if initErr != nil {
			return initErr
		}
		if g == nil {
			return errors.New("password Generator cannot be nil")
		}
		c.passGen = g
		return nil
	}
}

func WithNumGen(g SecureRandomByteser, intiErr error) Option {
	return func(c *authenticationConfig) error {
		if intiErr != nil {
			return intiErr
		}
		if g == nil {
			return errors.New("number Generator cannot be nil")
		}
		c.numGen = g
		return nil
	}
}

func WithURLTokenGen(g SecureRandomByteser, intiErr error) Option {
	return func(c *authenticationConfig) error {
		if intiErr != nil {
			return intiErr
		}
		if g == nil {
			return errors.New("URL token Generator cannot be nil")
		}
		c.urlTokenGen = g
		return nil
	}
}

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

func WithAppName(n string) Option {
	return func(c *authenticationConfig) error {
		c.appNameEmptyable = n
		return nil
	}
}

func WithFacebookCl(fb FacebookCl) Option {
	return func(c *authenticationConfig) error {
		c.fbNilable = fb
		return nil
	}
}

// WithSMSCl sets the SMS client to use.
// You may provide templates
// for sending SMSes e.g. WithPhoneVerifyTplt(), WithPhoneInviteTplt(),
// WithPhoneResetPassTplt() ...otherwise default template files in the config
// package are used.
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
// for sending Emails e.g. WithEmailVerifyTplt(), WithEmailInviteTplt(),
// WithEmailResetPassTplt() ...otherwise default template files in the config
// package are used.
// NewAuthentication() fails if this option is not nil and one of the
// templates couldn't be loaded.
func WithEmailCl(cl Mailer) Option {
	return func(c *authenticationConfig) error {
		c.mailerNilable = cl
		return nil
	}
}

func WithWebAppURL(URL string) Option {
	return func(c *authenticationConfig) error {
		c.webAppURLEmptyable = URL
		return nil
	}
}

func WithInvitationSubject(s string) Option {
	return func(c *authenticationConfig) error {
		c.invSubjEmptyable = s
		return nil
	}
}

func WithVerificationSubject(s string) Option {
	return func(c *authenticationConfig) error {
		c.verSubjEmptyable = s
		return nil
	}
}

func WithResetPassSubject(s string) Option {
	return func(c *authenticationConfig) error {
		c.resPassSubjEmptyable = s
		return nil
	}
}

func WithPhoneInviteTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil {
			return errors.New("provided phone invite template was nil")
		}
		c.loginTpActionTplts[loginTypePhone][ActionInvite] = t
		return nil
	}
}

func WithPhoneResetPassTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil {
			return errors.New("provided phone invite template was nil")
		}
		c.loginTpActionTplts[loginTypePhone][ActionResetPass] = t
		return nil
	}
}

func WithPhoneVerifyTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil {
			return errors.New("provided phone invite template was nil")
		}
		c.loginTpActionTplts[loginTypePhone][ActionVerify] = t
		return nil
	}
}

func WithEmailInviteTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil {
			return errors.New("provided phone invite template was nil")
		}
		c.loginTpActionTplts[loginTypeEmail][ActionInvite] = t
		return nil
	}
}

func WithEmailResetPassTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil {
			return errors.New("provided email reset pass template was nil")
		}
		c.loginTpActionTplts[loginTypeEmail][ActionResetPass] = t
		return nil
	}
}

func WithEmailVerifyTplt(t *template.Template, parseErr error) Option {
	return func(c *authenticationConfig) error {
		if parseErr != nil {
			return parseErr
		}
		if t == nil {
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
	// TODO get rid of this chunk, app deserves to panic if these conditions
	// aren't met. Reason for decision: c.initialzeValues() and c.fillDefaults()
	//
	//if c.passGen == nil {
	//	return errors.New("password generator was nil")
	//}
	//if c.numGen == nil {
	//	return errors.New("number generator was nil")
	//}
	//if c.urlTokenGen == nil {
	//	return errors.New("URL token generator was nil")
	//}
	//if c.loginTpActionTplts == nil {
	//	return errors.New("login type action templates maps were nil")
	//}
	//phoneActTpls, ok := c.loginTpActionTplts[loginTypePhone]
	//if !ok {
	//	return errors.New("phone action templates was nil")
	//}
	//if c.smserNilable != nil {
	//	if _, ok = phoneActTpls[ActionVerify]; !ok {
	//		return errors.New("phone verification template not found")
	//	}
	//	if _, ok = phoneActTpls[ActionInvite]; !ok {
	//		return errors.New("phone invite template not found")
	//	}
	//	if _, ok = phoneActTpls[ActionResetPass]; !ok {
	//		return errors.New("phone reset password template not found")
	//	}
	//}
	//emailActTpls, ok := c.loginTpActionTplts[loginTypeEmail]
	//if c.mailerNilable != nil {
	//	if _, ok = emailActTpls[ActionVerify]; !ok {
	//		return errors.New("email verification template not found")
	//	}
	//	if _, ok = emailActTpls[ActionInvite]; !ok {
	//		return errors.New("email invite template not found")
	//	}
	//	if _, ok = emailActTpls[ActionResetPass]; !ok {
	//		return errors.New("email reset password template not found")
	//	}
	//}
	//
	// End TODO

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
