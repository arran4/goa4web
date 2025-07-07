package main

import "flag"

// parseFlags builds a FlagSet with the provided name, applies flag
// registrations via fn, parses args and returns the FlagSet with any
// remaining positional arguments.
func parseFlags(name string, args []string, fn func(*flag.FlagSet)) (*flag.FlagSet, []string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	if fn != nil {
		fn(fs)
	}
	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}
	return fs, fs.Args(), nil
}
