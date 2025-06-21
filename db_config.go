package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/config"
)

// DBConfig holds parameters for connecting to the database.
type DBConfig struct {
	User         string
	Pass         string
	Host         string
	Port         string
	Name         string
	LogVerbosity int
}

// cliDBConfig is populated from command line flags in main().
var cliDBConfig DBConfig

// dbConfigFile is the optional path to a configuration file read at startup.
// If empty, the DB_CONFIG_FILE environment variable is consulted.
var dbConfigFile string

// resolveDBConfig merges configuration values with the order of precedence
// cli > file > env > defaults.
func resolveDBConfig(cli, file, env DBConfig) DBConfig {
	var cfg DBConfig
	config.Merge(&cfg, env)
	config.Merge(&cfg, file)
	config.Merge(&cfg, cli)
	return cfg
}

// loadDBConfigFile reads DB_* style configuration values from a simple key=value
// file. Missing files return an empty configuration.
func loadDBConfigFile(path string) (DBConfig, error) {
	var cfg DBConfig
	if path == "" {
		return cfg, nil
	}
	b, err := readFile(path)
	if err != nil {
		return cfg, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if i := strings.IndexByte(line, '='); i > 0 {
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			switch key {
			case config.EnvDBUser:
				cfg.User = val
			case config.EnvDBPass:
				cfg.Pass = val
			case config.EnvDBHost:
				cfg.Host = val
			case config.EnvDBPort:
				cfg.Port = val
			case config.EnvDBName:
				cfg.Name = val
			case "DB_LOG_VERBOSITY":
				if v, err := strconv.Atoi(val); err == nil {
					cfg.LogVerbosity = v
				}
			}
		}
	}
	return cfg, nil
}

// loadDBConfig loads the database configuration from environment, optional file
// and command line flags applying the precedence defined in AGENTS.md.
func loadDBConfig() DBConfig {
	env := DBConfig{
		User: os.Getenv(config.EnvDBUser),
		Pass: os.Getenv(config.EnvDBPass),
		Host: os.Getenv(config.EnvDBHost),
		Port: os.Getenv(config.EnvDBPort),
		Name: os.Getenv(config.EnvDBName),
	}
	if lv := os.Getenv("DB_LOG_VERBOSITY"); lv != "" {
		if v, err := strconv.Atoi(lv); err == nil {
			env.LogVerbosity = v
		}
	}

	cfgPath := dbConfigFile
	if cfgPath == "" {
		cfgPath = os.Getenv("DB_CONFIG_FILE")
	}
	fileCfg, err := loadDBConfigFile(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("DB config file error: %v", err)
	}

	return resolveDBConfig(cliDBConfig, fileCfg, env)
}
