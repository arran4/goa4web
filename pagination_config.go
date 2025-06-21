package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/config"
)

// PaginationConfig holds allowed minimum and maximum page sizes.
type PaginationConfig struct {
	Min int
	Max int
}

var cliPaginationConfig PaginationConfig
var paginationConfigFile string
var appPaginationConfig PaginationConfig

func resolvePaginationConfig(cli, file, env PaginationConfig) PaginationConfig {
	var cfg PaginationConfig
	if env.Min != 0 {
		cfg.Min = env.Min
	}
	if env.Max != 0 {
		cfg.Max = env.Max
	}
	if file.Min != 0 {
		cfg.Min = file.Min
	}
	if file.Max != 0 {
		cfg.Max = file.Max
	}
	if cli.Min != 0 {
		cfg.Min = cli.Min
	}
	if cli.Max != 0 {
		cfg.Max = cli.Max
	}
	if cfg.Min == 0 {
		cfg.Min = 5
	}
	if cfg.Max == 0 {
		cfg.Max = 50
	}
	if cfg.Min > cfg.Max {
		cfg.Min = cfg.Max
	}
	return cfg
}

func loadPaginationConfigFile(path string) (PaginationConfig, error) {
	var cfg PaginationConfig
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
			case "PAGE_SIZE_MIN":
				if v, err := strconv.Atoi(val); err == nil {
					cfg.Min = v
				}
			case "PAGE_SIZE_MAX":
				if v, err := strconv.Atoi(val); err == nil {
					cfg.Max = v
				}
			}
		}
	}
	return cfg, nil
}

func loadPaginationConfig() PaginationConfig {
	env := PaginationConfig{}
	if v := os.Getenv(config.EnvPageSizeMin); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			env.Min = n
		}
	}
	if v := os.Getenv(config.EnvPageSizeMax); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			env.Max = n
		}
	}

	cfgPath := paginationConfigFile
	if cfgPath == "" {
		cfgPath = os.Getenv("PAGINATION_CONFIG_FILE")
	}
	fileCfg, err := loadPaginationConfigFile(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("pagination config file error: %v", err)
	}

	appPaginationConfig = resolvePaginationConfig(cliPaginationConfig, fileCfg, env)
	return appPaginationConfig
}
