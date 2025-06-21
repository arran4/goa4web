package main

import (
	"log"
	"os"
	"strings"
)

// HTTPConfig holds parameters for the HTTP server.
type HTTPConfig struct {
	Listen   string
	Hostname string
}

// cliHTTPConfig is populated from command line flags in main().
var cliHTTPConfig HTTPConfig

// appHTTPConfig holds the resolved configuration after loadHTTPConfig is called.
var appHTTPConfig HTTPConfig

// httpConfigFile is the optional path to a configuration file read at startup.
var httpConfigFile string

// resolveHTTPConfig merges configuration values with the order of precedence
// cli > file > env > defaults.
func resolveHTTPConfig(cli, file, env HTTPConfig) HTTPConfig {
	var cfg HTTPConfig
	merge := func(src HTTPConfig) {
		if src.Listen != "" {
			cfg.Listen = src.Listen
		}
		if src.Hostname != "" {
			cfg.Hostname = src.Hostname
		}
	}
	merge(env)
	merge(file)
	merge(cli)
	if cfg.Listen == "" {
		cfg.Listen = ":8080"
	}
	if cfg.Hostname == "" {
		cfg.Hostname = "http://localhost:8080"
	}
	return cfg
}

// loadHTTPConfigFile reads LISTEN style configuration values from a simple
// key=value file. Missing files return an empty configuration.
func loadHTTPConfigFile(path string) (HTTPConfig, error) {
	var cfg HTTPConfig
	if path == "" {
		return cfg, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if i := strings.IndexByte(line, '='); i > 0 {
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			switch key {
			case "LISTEN":
				cfg.Listen = val
			case "HOSTNAME":
				cfg.Hostname = val
			}
		}
	}
	return cfg, nil
}

// loadHTTPConfig loads the HTTP configuration from environment, optional file
// and command line flags applying the precedence defined in AGENTS.md.
func loadHTTPConfig() HTTPConfig {
	env := HTTPConfig{
		Listen:   os.Getenv("LISTEN"),
		Hostname: os.Getenv("HOSTNAME"),
	}
	fileCfg, err := loadHTTPConfigFile(httpConfigFile)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("HTTP config file error: %v", err)
	}
	appHTTPConfig = resolveHTTPConfig(cliHTTPConfig, fileCfg, env)
	return appHTTPConfig
}
