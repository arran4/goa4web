package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// DefaultPageSize defines the number of items shown per page.
const DefaultPageSize = 15

// RuntimeConfig stores configuration values resolved from environment
// variables, optional files and command line flags.
type RuntimeConfig struct {
	DBConn            string
	DBDriver          string
	DBTimezone        string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPass            string
	DBName            string
	DBLogVerbosity    int
	EmailLogVerbosity int
	LogFlags          int

	HTTPListen   string
	HTTPHostname string
	// HSTSHeaderValue defines the Strict-Transport-Security header value.
	HSTSHeaderValue string

	EmailProvider             string
	EmailSMTPHost             string
	EmailSMTPPort             string
	EmailSMTPUser             string
	EmailSMTPPass             string
	EmailSMTPAuth             string
	EmailSMTPStartTLS         bool
	EmailFrom                 string
	EmailAWSRegion            string
	EmailJMAPEndpoint         string
	EmailJMAPEndpointOverride string
	EmailJMAPAccount          string
	EmailJMAPIdentity         string
	EmailJMAPUser             string
	EmailJMAPPass             string
	EmailJMAPInsecure         bool
	EmailSendGridKey          string
	EmailSubjectPrefix        string
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
	// EmailVerificationExpiryHours sets how long verification links remain valid.
	EmailVerificationExpiryHours int
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
	// Timezone defines the default site timezone used when users have not
	// specified their own.
	Timezone string

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

	// ShareSignSecret is used to sign share URLs.
	ShareSignSecret string
	// ShareSignSecretFile specifies the path to the share signing key.
	ShareSignSecretFile string

	// AdminAPISecret is used to sign administrator API tokens.
	AdminAPISecret string
	// AdminAPISecretFile specifies the path to the administrator API signing key.
	AdminAPISecretFile string

	// TemplatesDir specifies a directory to load templates and assets from.
	TemplatesDir string
	// AutoMigrate toggles automatic database migrations on startup.
	AutoMigrate bool
	// MigrationsDir specifies a directory to load migrations from at runtime.
	MigrationsDir string

	// CreateDirs creates missing directories when enabled.
	CreateDirs bool

	// OGImageWidth is the width of the generated Open Graph image.
	OGImageWidth int
	// OGImageHeight is the height of the generated Open Graph image.
	OGImageHeight int
	// OGImagePattern is the pattern style to use for the Open Graph image.
	OGImagePattern string
	// OGImageFgColor is the foreground color for the Open Graph image.
	OGImageFgColor string
	// OGImageBgColor is the background color for the Open Graph image.
	OGImageBgColor string
	// OGImageRpgTheme uses the RPG theme for the Open Graph image.
	OGImageRpgTheme bool
	// TwitterSite is the Twitter handle for the site.
	TwitterSite string
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
	if cfg.DBConn != "" {
		if cfg.DBUser != "" || cfg.DBPass != "" || cfg.DBHost != "" || cfg.DBPort != "" {
			conf, err := mysql.ParseDSN(cfg.DBConn)
			if err != nil {
				log.Fatalf("Invalid DB_CONN: %v", err)
			}
			if cfg.DBUser != "" && conf.User != cfg.DBUser {
				log.Fatalf("DB_CONN user (%s) contradicts DB_USER (%s)", conf.User, cfg.DBUser)
			}
			if cfg.DBPass != "" && conf.Passwd != cfg.DBPass {
				log.Fatalf("DB_CONN password contradicts DB_PASS")
			}
			// DB_NAME check
			if cfg.DBName != "" && conf.DBName != cfg.DBName {
				log.Fatalf("DB_CONN database name (%s) contradicts DB_NAME (%s)", conf.DBName, cfg.DBName)
			}

			if (cfg.DBHost != "" || cfg.DBPort != "") && (conf.Net == "tcp" || conf.Net == "") {
				// Parse Addr which is host:port
				addr := conf.Addr
				if addr == "" {
					addr = "127.0.0.1:3306"
				}

				// Very basic parsing, might need 'net.SplitHostPort'
				host, port, err := splitHostPort(addr)
				if err == nil {
					if cfg.DBHost != "" && host != cfg.DBHost {
						log.Fatalf("DB_CONN host (%s) contradicts DB_HOST (%s)", host, cfg.DBHost)
					}
					if cfg.DBPort != "" && port != cfg.DBPort {
						log.Fatalf("DB_CONN port (%s) contradicts DB_PORT (%s)", port, cfg.DBPort)
					}
				}
			}
		}
	} else if cfg.DBHost != "" || cfg.DBUser != "" || cfg.DBName != "" {
		// Construct DB_CONN from components if DB_CONN is missing
		// Default to tcp
		user := cfg.DBUser
		pass := cfg.DBPass
		host := cfg.DBHost
		port := cfg.DBPort
		dbname := cfg.DBName

		if host == "" {
			host = "127.0.0.1"
		}
		if port == "" {
			port = "3306"
		}

		// Format: user:password@tcp(host:port)/dbname
		// Need to handle missing user/pass
		auth := ""
		if user != "" {
			auth = user
			if pass != "" {
				auth += ":" + pass
			}
			auth += "@"
		}

		cfg.DBConn = fmt.Sprintf("%stcp(%s:%s)/%s?parseTime=true", auth, host, port, dbname)
	}

	if cfg.HTTPHostname == "" {
		cfg.HTTPHostname = "http://localhost:8080"
	}
	cfg.HTTPHostname = strings.TrimSuffix(cfg.HTTPHostname, "/")
	if cfg.PageSizeMin > cfg.PageSizeMax {
		cfg.PageSizeMin = cfg.PageSizeMax
	}
	if cfg.PageSizeDefault < cfg.PageSizeMin {
		cfg.PageSizeDefault = cfg.PageSizeMin
	}
	if cfg.PageSizeDefault > cfg.PageSizeMax {
		cfg.PageSizeDefault = cfg.PageSizeMax
	}
	if cfg.ImageUploadDir == "" {
		if os.Getenv(EnvDocker) != "" {
			cfg.ImageUploadDir = "/var/lib/goa4web/images"
		} else {
			cfg.ImageUploadDir = "uploads/images"
		}
	}
	if cfg.ImageCacheDir == "" {
		if os.Getenv(EnvDocker) != "" {
			cfg.ImageCacheDir = "/var/cache/goa4web/thumbnails"
		} else {
			cfg.ImageCacheDir = "uploads/cache"
		}
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
	if cfg.EmailVerificationExpiryHours == 0 {
		cfg.EmailVerificationExpiryHours = 24
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

func splitHostPort(addr string) (string, string, error) {
	return net.SplitHostPort(addr)
}
