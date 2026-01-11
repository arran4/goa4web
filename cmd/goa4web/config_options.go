package main

import (
	"bytes"
	"flag"
	"sort"

	"github.com/arran4/goa4web/config"
)

// configOptionsCmd implements "config options".
type configOptionsCmd struct {
	*configCmd
	fs       *flag.FlagSet
	template string
	extended bool
}

func parseConfigOptionsCmd(parent *configCmd, args []string) (*configOptionsCmd, error) {
	c := &configOptionsCmd{configCmd: parent}
	c.fs = newFlagSet("options")
	c.fs.StringVar(&c.template, "template", "", "template file")
	c.fs.BoolVar(&c.extended, "extended", false, "include extended usage")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

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
	usage := config.UsageMap()
	ext := config.ExtendedUsageMap(c.rootCmd.dbReg)
	names := config.NameMap()
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
	return executeUsage(bytes.NewBuffer(nil), "config_options.txt", c)
}

func (c *configOptionsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*configOptionsCmd)(nil)
