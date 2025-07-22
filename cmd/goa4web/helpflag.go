package main

import "flag"

// usageIfHelp prints usage if the first argument is a help keyword.
// It returns flag.ErrHelp when usage was shown.
func usageIfHelp(fs *flag.FlagSet, args []string) error {
	if len(args) > 0 {
		if args[0] == "help" || args[0] == "usage" {
			fs.Usage()
			return flag.ErrHelp
		}
	}
	return nil
}
