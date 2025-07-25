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
	DBConn         string
	DBDriver       string
	DBLogVerbosity int
	LogFlags       int

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
	// ImageSignSecret is used to sign image URLs.
	ImageSignSecret string
	// ImageSignSecretFile specifies the path to the image signing key.
	ImageSignSecretFile string

	// AdminAPISecret is used to sign administrator API tokens.
	AdminAPISecret string
	// AdminAPISecretFile specifies the path to the administrator API signing key.
	AdminAPISecretFile string

	// CreateDirs creates missing directories when enabled.
	CreateDirs bool
}

// AppRuntimeConfig stores the current application configuration.
var AppRuntimeConfig RuntimeConfig

// newRuntimeFlagSet returns a FlagSet containing the provided options merged
// with the built-in ones.
func newRuntimeFlagSet(name string, sopts []StringOption, iopts []IntOption, bopts []BoolOption) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)

	strings := append(append([]StringOption(nil), StringOptions...), sopts...)
	ints := append(append([]IntOption(nil), IntOptions...), iopts...)
	bools := append(append([]BoolOption(nil), BoolOptions...), bopts...)

	for _, o := range strings {
		fs.String(o.Name, o.Default, o.Usage)
	}
	for _, o := range ints {
		fs.Int(o.Name, o.Default, o.Usage)
	}
	for _, o := range bools {
		if o.Name != "" {
			fs.String(o.Name, "", o.Usage)
		}
	}

	return fs
}

// NewRuntimeFlagSet returns a FlagSet containing all runtime configuration
// options. The returned FlagSet uses ContinueOnError to allow tests to supply
// their own arguments without triggering os.Exit.
func NewRuntimeFlagSet(name string) *flag.FlagSet {
	return newRuntimeFlagSet(name, nil, nil, nil)
}

// NewRuntimeFlagSetWithOptions returns a FlagSet containing the built-in options
// merged with the provided slices.
func NewRuntimeFlagSetWithOptions(name string, sopts []StringOption, iopts []IntOption) *flag.FlagSet {
	return newRuntimeFlagSet(name, sopts, iopts, nil)
}

// generateRuntimeConfig constructs the RuntimeConfig from the provided option
// slices applying the standard precedence order of command line flags,
// configuration file values and environment variables.
func generateRuntimeConfig(fs *flag.FlagSet, fileVals map[string]string, getenv func(string) string, sopts []StringOption, iopts []IntOption, bopts []BoolOption) RuntimeConfig {
	if getenv == nil {
		getenv = os.Getenv
	}
	setFlags := map[string]bool{}
	if fs != nil {
		fs.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })
	}

	cfg := RuntimeConfig{}

	strings := append(append([]StringOption(nil), StringOptions...), sopts...)
	ints := append(append([]IntOption(nil), IntOptions...), iopts...)
	bools := append(append([]BoolOption(nil), BoolOptions...), bopts...)

	for _, o := range strings {
		dst := o.Target(&cfg)
		var val string
		if fs != nil && setFlags[o.Name] {
			if f := fs.Lookup(o.Name); f != nil {
				val = f.Value.String()
			}
		}
		if val == "" {
			val = fileVals[o.Env]
		}
		if val == "" {
			val = getenv(o.Env)
		}
		if val == "" {
			val = o.Default
		}
		*dst = val
	}

	for _, o := range ints {
		dst := o.Target(&cfg)
		var val string
		if fs != nil && setFlags[o.Name] {
			if f := fs.Lookup(o.Name); f != nil {
				val = f.Value.String()
			}
		}
		if val == "" {
			val = fileVals[o.Env]
		}
		if val == "" {
			val = getenv(o.Env)
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
		dst := o.Target(&cfg)
		var cliVal string
		if fs != nil && o.Name != "" && setFlags[o.Name] {
			if f := fs.Lookup(o.Name); f != nil {
				cliVal = f.Value.String()
			}
		}
		*dst = resolveBool(o.Default, cliVal, fileVals[o.Env], getenv(o.Env))
	}

	if cfg.SessionSecretFile == "" {
		cfg.SessionSecretFile = DefaultSessionSecretPath()
	}
	if cfg.AdminAPISecretFile == "" {
		cfg.AdminAPISecretFile = DefaultAdminAPISecretPath()
	}

	normalizeRuntimeConfig(&cfg)
	AppRuntimeConfig = cfg
	return cfg
}

// GenerateRuntimeConfig builds the runtime configuration from a FlagSet,
// optional config file values and environment variables following the
// precedence rules from AGENTS.md.
// GenerateRuntimeConfig builds the runtime configuration from a FlagSet,
// optional config file values and environment variables following the
// precedence rules from AGENTS.md. The getenv function is used to
// retrieve environment values and defaults to os.Getenv when nil.
func GenerateRuntimeConfig(fs *flag.FlagSet, fileVals map[string]string, getenv func(string) string) RuntimeConfig {
	return generateRuntimeConfig(fs, fileVals, getenv, nil, nil, nil)
}

// GenerateRuntimeConfigWithOptions is like GenerateRuntimeConfig but also considers
// the supplied option slices.
func GenerateRuntimeConfigWithOptions(fs *flag.FlagSet, fileVals map[string]string, getenv func(string) string, sopts []StringOption, iopts []IntOption) RuntimeConfig {
	return generateRuntimeConfig(fs, fileVals, getenv, sopts, iopts, nil)
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
