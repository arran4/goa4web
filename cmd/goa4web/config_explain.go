package main

import (
	"flag"
	"fmt"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/configexplain"
)

type configExplainCmd struct {
	*configCmd
	fs *flag.FlagSet
}

func parseConfigExplainCmd(parent *configCmd, args []string) (*configExplainCmd, error) {
	c := &configExplainCmd{configCmd: parent}
	c.fs = newFlagSet("explain")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *configExplainCmd) Run() error {
	args := c.fs.Args()
	if len(args) > 0 && args[0] != "source" {
		c.fs.Usage()
		return fmt.Errorf("unknown explain command %q", args[0])
	}

	fileVals := c.rootCmd.ConfigFileValues

	w := tabwriter.NewWriter(c.fs.Output(), 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "Option\tFinal Value\tSource\tDetail")

	infos := configexplain.Explain(configexplain.Inputs{
		FlagSet:    c.rootCmd.fs,
		FileValues: fileVals,
		ConfigFile: c.ConfigFile,
	})

	for _, info := range infos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", info.Name, info.FinalValue, info.SourceLabel, info.SourceDetail)
	}

	w.Flush()
	return nil
}

func (c *configExplainCmd) Usage() {
	executeUsage(c.fs.Output(), "config_explain_usage.txt", c)
}

func (c *configExplainCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*configExplainCmd)(nil)
