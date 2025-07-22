package main

import (
	"flag"
	"fmt"
	"io"
	"text/template"
)

// newFlagSet returns a FlagSet preconfigured to print flags using the
// standard template.
func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = func() { printFlags(fs) }
	return fs
}

// parseFlags builds a FlagSet with the provided name, applies flag
// registrations via fn, parses args and returns the FlagSet with any
// remaining positional arguments.
func parseFlags(name string, args []string, fn func(*flag.FlagSet)) (*flag.FlagSet, []string, error) {
	fs := newFlagSet(name)
	if fn != nil {
		fn(fs)
	}
	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}
	return fs, fs.Args(), nil
}

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
	t := template.Must(template.New("flags").Parse(templateString("flags.txt")))
	if err := t.Execute(fs.Output(), flagInfos(fs)); err != nil {
		fmt.Fprintf(fs.Output(), "template execute: %v\n", err)
	}
}

func executeUsage(w io.Writer, tmplStr string, fs *flag.FlagSet, prog string) {
	t := template.Must(template.New("usage").Parse(tmplStr))
	t = template.Must(t.New("flags").Parse(templateString("flags.txt")))
	if err := t.Execute(w, struct {
		Prog  string
		Flags []flagInfo
	}{Prog: prog, Flags: flagInfos(fs)}); err != nil {
		fmt.Fprintf(w, "template execute: %v\n", err)
	}
}
