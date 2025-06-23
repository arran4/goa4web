package main

import "strconv"

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
