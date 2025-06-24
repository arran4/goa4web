package runtimeconfig

import (
	"flag"
	"os"
	"strconv"

	"github.com/arran4/goa4web/config"
)

// DefaultPageSize defines the number of items shown per page.
const DefaultPageSize = 15

// RuntimeConfig stores configuration values resolved from environment
// variables, optional files and command line flags.
type RuntimeConfig struct {
	DBUser         string
	DBPass         string
	DBHost         string
	DBPort         string
	DBName         string
	DBLogVerbosity int

	HTTPListen   string
	HTTPHostname string

	EmailProvider     string
	EmailSMTPHost     string
	EmailSMTPPort     string
	EmailSMTPUser     string
	EmailSMTPPass     string
	EmailAWSRegion    string
	EmailJMAPEndpoint string
	EmailJMAPAccount  string
	EmailJMAPIdentity string
	EmailJMAPUser     string
	EmailJMAPPass     string
	EmailSendGridKey  string

	PageSizeMin     int
	PageSizeMax     int
	PageSizeDefault int

	FeedsEnabled    bool
	StatsStartYear  int
	DefaultLanguage string

	ImageUploadDir string
	ImageMaxBytes  int
}

// AppRuntimeConfig stores the current application configuration.
var AppRuntimeConfig RuntimeConfig

// NewRuntimeFlagSet returns a FlagSet containing all runtime configuration
// options. The returned FlagSet uses ContinueOnError to allow tests to supply
// their own arguments without triggering os.Exit.
func NewRuntimeFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)

	fs.String("db-user", "", "database user")
	fs.String("db-pass", "", "database password")
	fs.String("db-host", "", "database host")
	fs.String("db-port", "", "database port")
	fs.String("db-name", "", "database name")
	fs.Int("db-log-verbosity", 0, "database logging verbosity")

	fs.String("listen", ":8080", "server listen address")
	fs.String("hostname", "", "server base URL")

	fs.String("email-provider", "", "email provider")
	fs.String("smtp-host", "", "SMTP host")
	fs.String("smtp-port", "", "SMTP port")
	fs.String("smtp-user", "", "SMTP user")
	fs.String("smtp-pass", "", "SMTP pass")
	fs.String("aws-region", "", "AWS region")
	fs.String("jmap-endpoint", "", "JMAP endpoint")
	fs.String("jmap-account", "", "JMAP account")
	fs.String("jmap-identity", "", "JMAP identity")
	fs.String("jmap-user", "", "JMAP user")
	fs.String("jmap-pass", "", "JMAP pass")
	fs.String("sendgrid-key", "", "SendGrid API key")

	fs.Int("page-size-min", 0, "minimum allowed page size")
	fs.Int("page-size-max", 0, "maximum allowed page size")
	fs.Int("page-size-default", 0, "default page size")

	fs.String("feeds-enabled", "", "enable or disable feeds")
	fs.String("stats-start-year", "", "start year for usage stats")
	fs.String("default-language", "", "default language name")
	fs.String("image-upload-dir", "", "directory to store uploaded images")
	fs.Int("image-max-bytes", 0, "maximum allowed upload size in bytes")

	return fs
}

// GenerateRuntimeConfig builds the runtime configuration from a FlagSet,
// optional config file values and environment variables following the
// precedence rules from AGENTS.md.
func GenerateRuntimeConfig(fs *flag.FlagSet, fileVals map[string]string) RuntimeConfig {
	setFlags := map[string]bool{}
	if fs != nil {
		fs.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })
	}

	cfg := RuntimeConfig{}

	strOpts := []struct {
		name string
		env  string
		dst  *string
	}{
		{"db-user", config.EnvDBUser, &cfg.DBUser},
		{"db-pass", config.EnvDBPass, &cfg.DBPass},
		{"db-host", config.EnvDBHost, &cfg.DBHost},
		{"db-port", config.EnvDBPort, &cfg.DBPort},
		{"db-name", config.EnvDBName, &cfg.DBName},
		{"listen", config.EnvListen, &cfg.HTTPListen},
		{"hostname", config.EnvHostname, &cfg.HTTPHostname},
		{"email-provider", config.EnvEmailProvider, &cfg.EmailProvider},
		{"smtp-host", config.EnvSMTPHost, &cfg.EmailSMTPHost},
		{"smtp-port", config.EnvSMTPPort, &cfg.EmailSMTPPort},
		{"smtp-user", config.EnvSMTPUser, &cfg.EmailSMTPUser},
		{"smtp-pass", config.EnvSMTPPass, &cfg.EmailSMTPPass},
		{"aws-region", config.EnvAWSRegion, &cfg.EmailAWSRegion},
		{"jmap-endpoint", config.EnvJMAPEndpoint, &cfg.EmailJMAPEndpoint},
		{"jmap-account", config.EnvJMAPAccount, &cfg.EmailJMAPAccount},
		{"jmap-identity", config.EnvJMAPIdentity, &cfg.EmailJMAPIdentity},
		{"jmap-user", config.EnvJMAPUser, &cfg.EmailJMAPUser},
		{"jmap-pass", config.EnvJMAPPass, &cfg.EmailJMAPPass},
		{"sendgrid-key", config.EnvSendGridKey, &cfg.EmailSendGridKey},
		{"default-language", config.EnvDefaultLanguage, &cfg.DefaultLanguage},
		{"image-upload-dir", config.EnvImageUploadDir, &cfg.ImageUploadDir},
	}
	for _, o := range strOpts {
		if fs != nil && setFlags[o.name] {
			if f := fs.Lookup(o.name); f != nil {
				*o.dst = f.Value.String()
				continue
			}
		}
		if v := fileVals[o.env]; v != "" {
			*o.dst = v
			continue
		}
		if v := os.Getenv(o.env); v != "" {
			*o.dst = v
		}
	}

	intOpts := []struct {
		name string
		env  string
		dst  *int
	}{
		{"db-log-verbosity", config.EnvDBLogVerbosity, &cfg.DBLogVerbosity},
		{"page-size-min", config.EnvPageSizeMin, &cfg.PageSizeMin},
		{"page-size-max", config.EnvPageSizeMax, &cfg.PageSizeMax},
		{"page-size-default", config.EnvPageSizeDefault, &cfg.PageSizeDefault},
		{"image-max-bytes", config.EnvImageMaxBytes, &cfg.ImageMaxBytes},
	}
	for _, o := range intOpts {
		var val string
		if fs != nil && setFlags[o.name] {
			if f := fs.Lookup(o.name); f != nil {
				val = f.Value.String()
			}
		} else if v := fileVals[o.env]; v != "" {
			val = v
		} else if v := os.Getenv(o.env); v != "" {
			val = v
		}
		if val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				*o.dst = n
			}
		}
	}

	var cliFeeds, cliStats string
	if fs != nil && setFlags["feeds-enabled"] {
		cliFeeds = fs.Lookup("feeds-enabled").Value.String()
	}
	if fs != nil && setFlags["stats-start-year"] {
		cliStats = fs.Lookup("stats-start-year").Value.String()
	}

	cfg.FeedsEnabled = resolveFeedsEnabled(
		cliFeeds,
		fileVals[config.EnvFeedsEnabled],
		os.Getenv(config.EnvFeedsEnabled),
	)
	cfg.StatsStartYear = resolveStatsStartYear(
		cliStats,
		fileVals[config.EnvStatsStartYear],
		os.Getenv(config.EnvStatsStartYear),
	)

	normalizeRuntimeConfig(&cfg)
	AppRuntimeConfig = cfg
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
	if cfg.ImageUploadDir == "" {
		cfg.ImageUploadDir = "uploads/images"
	}
	if cfg.ImageMaxBytes == 0 {
		cfg.ImageMaxBytes = 5 * 1024 * 1024
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
