package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"sort"
	"text/template"

	"github.com/arran4/goa4web/runtimeconfig"
)

//go:embed templates/config_options.txt
var configOptionsDefaultTemplate string

// configOptionsCmd implements "config options".
type configOptionsCmd struct {
	*configCmd
	fs       *flag.FlagSet
	template string
	extended bool
	args     []string
}

func parseConfigOptionsCmd(parent *configCmd, args []string) (*configOptionsCmd, error) {
	c := &configOptionsCmd{configCmd: parent}
	fs := flag.NewFlagSet("options", flag.ContinueOnError)
	fs.StringVar(&c.template, "template", "", "template file")
	fs.BoolVar(&c.extended, "extended", false, "include extended usage")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

type option struct {
	Env      string
	Flag     string
	Default  string
	Usage    string
	Extended string
}

func (c *configOptionsCmd) Run() error {
	def := defaultMap()
	usage := runtimeconfig.UsageMap()
	ext := runtimeconfig.ExtendedUsageMap()
	names := runtimeconfig.NameMap()
	keys := make([]string, 0, len(def))
	for k := range def {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	opts := make([]option, 0, len(keys))
	for _, k := range keys {
		e := ""
		if c.extended {
			e = ext[k]
		}
		opts = append(opts, option{
			Env:      k,
			Flag:     names[k],
			Default:  def[k],
			Usage:    usage[k],
			Extended: e,
		})
	}
	tmplText := configOptionsDefaultTemplate
	if c.template != "" {
		b, err := os.ReadFile(c.template)
		if err != nil {
			return fmt.Errorf("read template: %w", err)
		}
		tmplText = string(b)
	}
	t, err := template.New("options").Parse(tmplText)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	return t.Execute(os.Stdout, opts)
}
