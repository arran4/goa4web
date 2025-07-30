package config

import (
	"flag"
	"os"
	"strconv"
)

// DefaultPageSize defines the number of items shown per page.
const DefaultPageSize = 15

// RuntimeConfig stores configuration values resolved from environment
// variables, optional files and command line flags.
type RuntimeConfig struct {
	DBConn            string
	DBDriver          string
	DBLogVerbosity    int
	EmailLogVerbosity int
	LogFlags          int

	HTTPListen   string
	HTTPHostname string
	// HSTSHeaderValue defines the Strict-Transport-Security header value.
	HSTSHeaderValue string

	EmailProvider      string
	EmailSMTPHost      string
	EmailSMTPPort      string
	EmailSMTPUser      string
	EmailSMTPPass      string
	EmailSMTPAuth      string
	EmailSMTPStartTLS  bool
	EmailFrom          string
	EmailAWSRegion     string
	EmailJMAPEndpoint  string
	EmailJMAPAccount   string
	EmailJMAPIdentity  string
	EmailJMAPUser      string
	EmailJMAPPass      string
	EmailSendGridKey   string
	EmailSubjectPrefix string
	// EmailSignOff defines the optional sign-off appended to emails.
	EmailSignOff string

	// EmailEnabled toggles sending queued emails.
	EmailEnabled bool
	// NotificationsEnabled toggles the internal notification system.
	NotificationsEnabled bool
	// CSRFEnabled enables or disables CSRF protection.
	CSRFEnabled bool
	// AdminNotify toggles administrator notification emails.
	AdminNotify bool

	// AdminEmails holds a comma-separated list of administrator email
	// addresses.
	AdminEmails string

	// EmailWorkerInterval sets how often the email worker runs in seconds.
	EmailWorkerInterval int
	// PasswordResetExpiryHours sets how long password reset requests remain valid.
	PasswordResetExpiryHours int
	// LoginAttemptWindow defines the timeframe in minutes used when counting
	// failed login attempts for throttling.
	LoginAttemptWindow int
	// LoginAttemptThreshold is the maximum number of failed login attempts
	// allowed within the window.
	LoginAttemptThreshold int

	PageSizeMin     int
	PageSizeMax     int
	PageSizeDefault int

	FeedsEnabled    bool
	StatsStartYear  int
	DefaultLanguage string

	ImageUploadProvider string
	ImageUploadDir      string
	ImageUploadS3URL    string
	ImageCacheProvider  string
	ImageCacheDir       string
	ImageCacheS3URL     string
	ImageMaxBytes       int
	ImageCacheMaxBytes  int

	DLQProvider string
	DLQFile     string

	// SessionName specifies the cookie name used for session data.
	SessionName string

	// SessionSecret holds the session secret used to encrypt cookies.
	SessionSecret string
	// SessionSecretFile specifies the path to the session secret file.
	SessionSecretFile string
	// SessionSameSite selects the cookie SameSite mode used for sessions.
	SessionSameSite string
	// ImageSignSecret is used to sign image URLs.
	ImageSignSecret string
	// ImageSignSecretFile specifies the path to the image signing key.
	ImageSignSecretFile string

	// LinkSignSecret is used to sign external link URLs.
	LinkSignSecret string
	// LinkSignSecretFile specifies the path to the external link signing key.
	LinkSignSecretFile string

	// AdminAPISecret is used to sign administrator API tokens.
	AdminAPISecret string
	// AdminAPISecretFile specifies the path to the administrator API signing key.
	AdminAPISecretFile string

	// CreateDirs creates missing directories when enabled.
	CreateDirs bool
}

// Option configures RuntimeConfig values.
// Option adjusts how a RuntimeConfig is generated or modifies the resulting
// configuration.
type Option func(*runtimeArgs)

type runtimeArgs struct {
	fs       *flag.FlagSet
	fileVals map[string]string
	getenv   func(string) string
	sopts    []StringOption
	iopts    []IntOption
	bopts    []BoolOption
	post     []func(*RuntimeConfig)
}

// WithFlagSet supplies command line flags for configuration values.
func WithFlagSet(fs *flag.FlagSet) Option { return func(a *runtimeArgs) { a.fs = fs } }

// WithFileValues supplies configuration values loaded from a config file.
func WithFileValues(vals map[string]string) Option { return func(a *runtimeArgs) { a.fileVals = vals } }

// WithGetenv specifies the function used to lookup environment variables.
func WithGetenv(fn func(string) string) Option { return func(a *runtimeArgs) { a.getenv = fn } }

// WithStringOptions adds custom StringOptions used when parsing values.
func WithStringOptions(opts []StringOption) Option {
	return func(a *runtimeArgs) { a.sopts = append(a.sopts, opts...) }
}

// WithIntOptions adds custom IntOptions used when parsing values.
func WithIntOptions(opts []IntOption) Option {
	return func(a *runtimeArgs) { a.iopts = append(a.iopts, opts...) }
}

// WithBoolOptions adds custom BoolOptions used when parsing values.
func WithBoolOptions(opts []BoolOption) Option {
	return func(a *runtimeArgs) { a.bopts = append(a.bopts, opts...) }
}

// WithRuntimeConfig allows post-processing of the generated configuration.
func WithRuntimeConfig(fn func(*RuntimeConfig)) Option {
	return func(a *runtimeArgs) { a.post = append(a.post, fn) }
}

// NewRuntimeFlagSet returns a FlagSet containing all runtime configuration
// options. The returned FlagSet uses ContinueOnError to allow tests to supply
// their own arguments without triggering os.Exit.
func NewRuntimeFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, o := range StringOptions {
		fs.String(o.Name, o.Default, o.Usage)
	}
	for _, o := range IntOptions {
		fs.Int(o.Name, o.Default, o.Usage)
	}
	for _, o := range BoolOptions {
		if o.Name != "" {
			fs.String(o.Name, "", o.Usage)
		}
	}

	return fs
}

// NewRuntimeConfig constructs the runtime configuration by merging command line
// flags, values loaded from a config file and environment variables. Options can
// supply custom flag sets, configuration maps or environment lookup functions
// and may modify the resulting RuntimeConfig after it has been built.
func NewRuntimeConfig(ops ...Option) *RuntimeConfig {
	args := runtimeArgs{fileVals: map[string]string{}, getenv: os.Getenv}
	for _, op := range ops {
		op(&args)
	}
	if args.getenv == nil {
		args.getenv = os.Getenv
	}
	setFlags := map[string]bool{}
	if args.fs != nil {
		args.fs.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })
	}

	cfg := &RuntimeConfig{}

	strings := append(append([]StringOption(nil), StringOptions...), args.sopts...)
	ints := append(append([]IntOption(nil), IntOptions...), args.iopts...)
	bools := append(append([]BoolOption(nil), BoolOptions...), args.bopts...)

	for _, o := range strings {
		dst := o.Target(cfg)
		var val string
		if args.fs != nil && setFlags[o.Name] {
			if f := args.fs.Lookup(o.Name); f != nil {
				val = f.Value.String()
			}
		}
		if val == "" {
			val = args.fileVals[o.Env]
		}
		if val == "" {
			val = args.getenv(o.Env)
		}
		if val == "" {
			val = o.Default
		}
		*dst = val
	}

	for _, o := range ints {
		dst := o.Target(cfg)
		var val string
		if args.fs != nil && setFlags[o.Name] {
			if f := args.fs.Lookup(o.Name); f != nil {
				val = f.Value.String()
			}
		}
		if val == "" {
			val = args.fileVals[o.Env]
		}
		if val == "" {
			val = args.getenv(o.Env)
		}
		if val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				*dst = n
			}
		} else if o.Default != 0 {
			*dst = o.Default
		}
	}

	for _, o := range bools {
		dst := o.Target(cfg)
		var cliVal string
		if args.fs != nil && o.Name != "" && setFlags[o.Name] {
			if f := args.fs.Lookup(o.Name); f != nil {
				cliVal = f.Value.String()
			}
		}
		*dst = resolveBool(o.Default, cliVal, args.fileVals[o.Env], args.getenv(o.Env))
	}

	if cfg.SessionSecretFile == "" {
		cfg.SessionSecretFile = DefaultSessionSecretPath()
	}
	if cfg.AdminAPISecretFile == "" {
		cfg.AdminAPISecretFile = DefaultAdminAPISecretPath()
	}

	normalizeRuntimeConfig(cfg)
	for _, fn := range args.post {
		fn(cfg)
	}
	return cfg
}

// normalizeRuntimeConfig applies default values and ensures pagination limits are valid.
func normalizeRuntimeConfig(cfg *RuntimeConfig) {
	if cfg.HTTPListen == "" {
		cfg.HTTPListen = ":8080"
	}
	if cfg.HTTPHostname == "" {
		cfg.HTTPHostname = "http://localhost:8080"
	}
	if cfg.HSTSHeaderValue == "" {
		cfg.HSTSHeaderValue = "max-age=63072000; includeSubDomains"
	}
	if cfg.SessionName == "" {
		cfg.SessionName = "my-session"
	}
	if cfg.SessionSameSite == "" {
		cfg.SessionSameSite = "strict"
	}
	if cfg.PageSizeMin == 0 {
		cfg.PageSizeMin = 5
	}
	if cfg.PageSizeMax == 0 {
		cfg.PageSizeMax = 50
	}
	if cfg.PageSizeDefault == 0 {
		cfg.PageSizeDefault = DefaultPageSize
	}
	if cfg.PageSizeMin > cfg.PageSizeMax {
		cfg.PageSizeMin = cfg.PageSizeMax
	}
	if cfg.PageSizeDefault < cfg.PageSizeMin {
		cfg.PageSizeDefault = cfg.PageSizeMin
	}
	if cfg.PageSizeDefault > cfg.PageSizeMax {
		cfg.PageSizeDefault = cfg.PageSizeMax
	}
	if cfg.StatsStartYear == 0 {
		cfg.StatsStartYear = 2005
	}
	if cfg.ImageUploadProvider == "" {
		cfg.ImageUploadProvider = "local"
	}
	if cfg.ImageUploadDir == "" {
		cfg.ImageUploadDir = "uploads/images"
	}
	if cfg.ImageCacheProvider == "" {
		cfg.ImageCacheProvider = "local"
	}
	if cfg.ImageCacheDir == "" {
		cfg.ImageCacheDir = "uploads/cache"
	}
	if cfg.ImageMaxBytes == 0 {
		cfg.ImageMaxBytes = 50 * 1024 * 1024
	}
	if cfg.EmailSMTPAuth == "" {
		cfg.EmailSMTPAuth = "plain"
	}
	if cfg.EmailWorkerInterval == 0 {
		cfg.EmailWorkerInterval = 60
	}
	if cfg.PasswordResetExpiryHours == 0 {
		cfg.PasswordResetExpiryHours = 24
	}
	if cfg.LoginAttemptWindow == 0 {
		cfg.LoginAttemptWindow = 15
	}
	if cfg.LoginAttemptThreshold == 0 {
		cfg.LoginAttemptThreshold = 5
	}
}

// UpdatePaginationConfig adjusts the pagination fields of cfg and enforces
// valid limits.
func UpdatePaginationConfig(cfg *RuntimeConfig, min, max, def int) {
	if min != 0 {
		cfg.PageSizeMin = min
	}
	if max != 0 {
		cfg.PageSizeMax = max
	}
	if def != 0 {
		cfg.PageSizeDefault = def
	}
	normalizeRuntimeConfig(cfg)
}
