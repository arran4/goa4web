package main

import (
	"os"
	"strconv"

	"github.com/arran4/goa4web/config"
)

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
}

var cliRuntimeConfig RuntimeConfig
var appRuntimeConfig RuntimeConfig

// loadRuntimeConfig builds the runtime configuration from CLI flags, optional
// config files and environment variables following the precedence rules from
// AGENTS.md.
func loadRuntimeConfig(fileVals map[string]string) RuntimeConfig {
	env := RuntimeConfig{
		DBUser:            os.Getenv(config.EnvDBUser),
		DBPass:            os.Getenv(config.EnvDBPass),
		DBHost:            os.Getenv(config.EnvDBHost),
		DBPort:            os.Getenv(config.EnvDBPort),
		DBName:            os.Getenv(config.EnvDBName),
		HTTPListen:        os.Getenv(config.EnvListen),
		HTTPHostname:      os.Getenv(config.EnvHostname),
		EmailProvider:     os.Getenv(config.EnvEmailProvider),
		EmailSMTPHost:     os.Getenv(config.EnvSMTPHost),
		EmailSMTPPort:     os.Getenv(config.EnvSMTPPort),
		EmailSMTPUser:     os.Getenv(config.EnvSMTPUser),
		EmailSMTPPass:     os.Getenv(config.EnvSMTPPass),
		EmailAWSRegion:    os.Getenv(config.EnvAWSRegion),
		EmailJMAPEndpoint: os.Getenv(config.EnvJMAPEndpoint),
		EmailJMAPAccount:  os.Getenv(config.EnvJMAPAccount),
		EmailJMAPIdentity: os.Getenv(config.EnvJMAPIdentity),
		EmailJMAPUser:     os.Getenv(config.EnvJMAPUser),
		EmailJMAPPass:     os.Getenv(config.EnvJMAPPass),
		EmailSendGridKey:  os.Getenv(config.EnvSendGridKey),
	}
	if v := os.Getenv(config.EnvDBLogVerbosity); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			env.DBLogVerbosity = n
		}
	}
	if v := os.Getenv(config.EnvPageSizeMin); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			env.PageSizeMin = n
		}
	}
	if v := os.Getenv(config.EnvPageSizeMax); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			env.PageSizeMax = n
		}
	}
	if v := os.Getenv(config.EnvPageSizeDefault); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			env.PageSizeDefault = n
		}
	}

	fileCfg := RuntimeConfig{
		DBUser:            fileVals[config.EnvDBUser],
		DBPass:            fileVals[config.EnvDBPass],
		DBHost:            fileVals[config.EnvDBHost],
		DBPort:            fileVals[config.EnvDBPort],
		DBName:            fileVals[config.EnvDBName],
		HTTPListen:        fileVals[config.EnvListen],
		HTTPHostname:      fileVals[config.EnvHostname],
		EmailProvider:     fileVals[config.EnvEmailProvider],
		EmailSMTPHost:     fileVals[config.EnvSMTPHost],
		EmailSMTPPort:     fileVals[config.EnvSMTPPort],
		EmailSMTPUser:     fileVals[config.EnvSMTPUser],
		EmailSMTPPass:     fileVals[config.EnvSMTPPass],
		EmailAWSRegion:    fileVals[config.EnvAWSRegion],
		EmailJMAPEndpoint: fileVals[config.EnvJMAPEndpoint],
		EmailJMAPAccount:  fileVals[config.EnvJMAPAccount],
		EmailJMAPIdentity: fileVals[config.EnvJMAPIdentity],
		EmailJMAPUser:     fileVals[config.EnvJMAPUser],
		EmailJMAPPass:     fileVals[config.EnvJMAPPass],
		EmailSendGridKey:  fileVals[config.EnvSendGridKey],
	}
	if v := fileVals[config.EnvDBLogVerbosity]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			fileCfg.DBLogVerbosity = n
		}
	}
	if v := fileVals[config.EnvPageSizeMin]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			fileCfg.PageSizeMin = n
		}
	}
	if v := fileVals[config.EnvPageSizeMax]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			fileCfg.PageSizeMax = n
		}
	}
	if v := fileVals[config.EnvPageSizeDefault]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			fileCfg.PageSizeDefault = n
		}
	}

	cfg := RuntimeConfig{}
	config.Merge(&cfg, env)
	config.Merge(&cfg, fileCfg)
	config.Merge(&cfg, cliRuntimeConfig)

	normalizeRuntimeConfig(&cfg)
	appRuntimeConfig = cfg
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
}

// updatePaginationConfig adjusts the pagination fields of cfg and enforces
// valid limits.
func updatePaginationConfig(cfg *RuntimeConfig, min, max, def int) {
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
