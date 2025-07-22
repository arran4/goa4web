package main

import (
	"flag"
	"fmt"
	"io"
	fs2 "io/fs"
	"log"
	"sync"
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
	t := template.Must(template.New("flags").Parse("flags.txt"))
	if err := t.Execute(fs.Output(), flagInfos(fs)); err != nil {
		fmt.Fprintf(fs.Output(), "template execute: %v\n", err)
	}
}

var compiledTemplates *template.Template

func executeUsage(w io.Writer, filename string, fs *flag.FlagSet, prog string) error {
	sync.OnceFunc(func() {
		sub, err := fs2.Sub(templatesFS, "templates")
		if err != nil {
			log.Panicf("template sub err: %v", err)
		}
		compiledTemplates = template.Must(template.New("").ParseFS(sub, "*.txt"))
	})()
	type Data struct {
		Prog  string
		Flags []flagInfo
	}
	data := &Data{Prog: prog, Flags: flagInfos(fs)}
	if err := compiledTemplates.ExecuteTemplate(w, filename, data); err != nil {
		_, _ = fmt.Fprintf(w, "template execute: %v\n", err)
		return fmt.Errorf("execute template: %v", err)
	}
	return nil
}
