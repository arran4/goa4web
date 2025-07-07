package runtimeconfig

import (
	"flag"
	"os"
	"reflect"
	"strconv"

	"github.com/arran4/goa4web/config"
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

	EmailProvider     string
	EmailSMTPHost     string
	EmailSMTPPort     string
	EmailSMTPUser     string
	EmailSMTPPass     string
	EmailSMTPAuth     string
	EmailSMTPStartTLS bool
	EmailFrom         string
	EmailAWSRegion    string
	EmailJMAPEndpoint string
	EmailJMAPAccount  string
	EmailJMAPIdentity string
	EmailJMAPUser     string
	EmailJMAPPass     string
	EmailSendGridKey  string

	// AdminEmails holds a comma-separated list of administrator email
	// addresses.
	AdminEmails string

	// EmailWorkerInterval sets how often the email worker runs in seconds.
	EmailWorkerInterval int

	PageSizeMin     int
	PageSizeMax     int
	PageSizeDefault int

	FeedsEnabled    bool
	StatsStartYear  int
	DefaultLanguage string

	ImageUploadDir string
	ImageMaxBytes  int

	DLQProvider string
	DLQFile     string
}

// AppRuntimeConfig stores the current application configuration.
var AppRuntimeConfig RuntimeConfig

// newRuntimeFlagSet returns a FlagSet containing the provided options merged
// with the built-in ones.
func newRuntimeFlagSet(name string, sopts []StringOption, iopts []IntOption) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)

	strings := append(append([]StringOption(nil), StringOptions...), sopts...)
	ints := append(append([]IntOption(nil), IntOptions...), iopts...)

	for _, o := range strings {
		fs.String(o.Name, o.Default, o.Usage)
	}
	for _, o := range ints {
		fs.Int(o.Name, o.Default, o.Usage)
	}

	fs.String("feeds-enabled", "", "enable or disable feeds")
	fs.String("stats-start-year", "", "start year for usage stats")
	fs.String("smtp-starttls", "", "enable or disable STARTTLS")

	return fs
}

// NewRuntimeFlagSet returns a FlagSet containing all runtime configuration
// options. The returned FlagSet uses ContinueOnError to allow tests to supply
// their own arguments without triggering os.Exit.
func NewRuntimeFlagSet(name string) *flag.FlagSet {
	return newRuntimeFlagSet(name, nil, nil)
}

// NewRuntimeFlagSetWithOptions returns a FlagSet containing the built-in options
// merged with the provided slices.
func NewRuntimeFlagSetWithOptions(name string, sopts []StringOption, iopts []IntOption) *flag.FlagSet {
	return newRuntimeFlagSet(name, sopts, iopts)
}

// generateRuntimeConfig constructs the RuntimeConfig from the provided option
// slices applying the standard precedence order of command line flags,
// configuration file values and environment variables.
func generateRuntimeConfig(fs *flag.FlagSet, fileVals map[string]string, getenv func(string) string, sopts []StringOption, iopts []IntOption) RuntimeConfig {
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

	for _, o := range strings {
		dst := reflect.ValueOf(&cfg).Elem().FieldByName(o.Field)
		if !dst.IsValid() || dst.Kind() != reflect.String {
			continue
		}
		if fs != nil && setFlags[o.Name] {
			if f := fs.Lookup(o.Name); f != nil {
				dst.SetString(f.Value.String())
				continue
			}
		}
		if v := fileVals[o.Env]; v != "" {
			dst.SetString(v)
			continue
		}
		if v := getenv(o.Env); v != "" {
			dst.SetString(v)
		}
	}

	for _, o := range ints {
		dst := reflect.ValueOf(&cfg).Elem().FieldByName(o.Field)
		if !dst.IsValid() || dst.Kind() != reflect.Int {
			continue
		}
		var val string
		if fs != nil && setFlags[o.Name] {
			if f := fs.Lookup(o.Name); f != nil {
				val = f.Value.String()
			}
		} else if v := fileVals[o.Env]; v != "" {
			val = v
		} else if v := getenv(o.Env); v != "" {
			val = v
		}
		if val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				dst.SetInt(int64(n))
			}
		}
	}

	var cliFeeds, cliStats, cliStartTLS string
	if fs != nil && setFlags["feeds-enabled"] {
		cliFeeds = fs.Lookup("feeds-enabled").Value.String()
	}
	if fs != nil && setFlags["stats-start-year"] {
		cliStats = fs.Lookup("stats-start-year").Value.String()
	}
	if fs != nil && setFlags["smtp-starttls"] {
		cliStartTLS = fs.Lookup("smtp-starttls").Value.String()
	}

	cfg.FeedsEnabled = resolveFeedsEnabled(
		cliFeeds,
		fileVals[config.EnvFeedsEnabled],
		getenv(config.EnvFeedsEnabled),
	)
	cfg.StatsStartYear = resolveStatsStartYear(
		cliStats,
		fileVals[config.EnvStatsStartYear],
		getenv(config.EnvStatsStartYear),
	)
	cfg.EmailSMTPStartTLS = resolveSMTPStartTLS(
		cliStartTLS,
		fileVals[config.EnvSMTPStartTLS],
		getenv(config.EnvSMTPStartTLS),
	)

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
	return generateRuntimeConfig(fs, fileVals, getenv, nil, nil)
}

// GenerateRuntimeConfigWithOptions is like GenerateRuntimeConfig but also considers
// the supplied option slices.
func GenerateRuntimeConfigWithOptions(fs *flag.FlagSet, fileVals map[string]string, getenv func(string) string, sopts []StringOption, iopts []IntOption) RuntimeConfig {
	return generateRuntimeConfig(fs, fileVals, getenv, sopts, iopts)
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
	if cfg.EmailSMTPAuth == "" {
		cfg.EmailSMTPAuth = "plain"
	}
	if cfg.EmailWorkerInterval == 0 {
		cfg.EmailWorkerInterval = 60
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
