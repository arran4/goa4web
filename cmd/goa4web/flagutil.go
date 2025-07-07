package main

import (
	_ "embed"
	"flag"
	"io"
	"text/template"
)

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

//go:embed templates/flags.txt
var flagsTemplate string

type flagInfo struct {
	Name     string
	Usage    string
	DefValue string
}

func flagInfos(fs *flag.FlagSet) []flagInfo {
	var list []flagInfo
	fs.VisitAll(func(f *flag.Flag) {
		name, usage := flag.UnquoteUsage(f)
		list = append(list, flagInfo{Name: name, Usage: usage, DefValue: f.DefValue})
	})
	return list
}

func printFlags(fs *flag.FlagSet) {
	t := template.Must(template.New("flags").Parse(flagsTemplate))
	_ = t.Execute(fs.Output(), flagInfos(fs))
}

func executeUsage(w io.Writer, tmplStr string, fs *flag.FlagSet, prog string) {
	t := template.Must(template.New("usage").Parse(tmplStr))
	t = template.Must(t.New("flags").Parse(flagsTemplate))
	_ = t.Execute(w, struct {
		Prog  string
		Flags []flagInfo
	}{Prog: prog, Flags: flagInfos(fs)})
}
