package main

import (
	"os"
	"strconv"

	"github.com/arran4/goa4web/config"
)

var statsStartYear int

func parseInt(v string) (int, bool) {
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}

func resolveStatsStartYear(cli, file, env string) int {
	if n, ok := parseInt(cli); ok {
		return n
	}
	if n, ok := parseInt(file); ok {
		return n
	}
	if n, ok := parseInt(env); ok {
		return n
	}
	return 2005
}

func loadStatsStartYear(cli string, file map[string]string) {
	fileVal := ""
	if v, ok := file["STATS_START_YEAR"]; ok {
		fileVal = v
	}
	statsStartYear = resolveStatsStartYear(cli, fileVal, os.Getenv(config.EnvStatsStartYear))
}
