package configformat

import "flag"

// ParseAsFlags parses config "as-*" flags into an AsOptions value.
func ParseAsFlags(fs *flag.FlagSet, args []string) (AsOptions, error) {
	var opts AsOptions
	fs.BoolVar(&opts.Extended, "extended", false, "include extended usage")
	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	return opts, nil
}
