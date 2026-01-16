package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"

	"github.com/arran4/goa4web/config"
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

	setFlags := map[string]bool{}
	c.rootCmd.fs.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })

	fileVals := c.rootCmd.ConfigFileValues

	w := tabwriter.NewWriter(c.fs.Output(), 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "Option\tFinal Value\tSource\tDetail")

	type optionInfo struct {
		Name     string
		FinalVal string
		Source   string
		Detail   string
	}
	var infos []optionInfo

	for _, o := range config.StringOptions {
		finalVal := o.Default
		source := "Default"
		detail := ""

		flagVal := ""
		if f := c.rootCmd.fs.Lookup(o.Name); f != nil {
			flagVal = f.Value.String()
		}

		envVal := os.Getenv(o.Env)
		fileVal := fileVals[o.Env]

		if setFlags[o.Name] {
			finalVal = flagVal
			source = "Arg"
			detail = fmt.Sprintf("--%s", o.Name)
		} else if fileVal != "" {
			finalVal = fileVal
			source = "Config File"
			detail = fmt.Sprintf("%s key: %s", c.ConfigFile, o.Env)
		} else if envVal != "" {
			finalVal = envVal
			source = "Environment"
			detail = fmt.Sprintf("%s=%s", o.Env, envVal)
		}

		infos = append(infos, optionInfo{o.Name, finalVal, source, detail})
	}

	for _, o := range config.IntOptions {
		finalVal := strconv.Itoa(o.Default)
		source := "Default"
		detail := ""

		flagVal := ""
		if f := c.rootCmd.fs.Lookup(o.Name); f != nil {
			flagVal = f.Value.String()
		}

		envVal := os.Getenv(o.Env)
		fileVal := fileVals[o.Env]

		if setFlags[o.Name] {
			finalVal = flagVal
			source = "Arg"
			detail = fmt.Sprintf("--%s", o.Name)
		} else if fileVal != "" {
			finalVal = fileVal
			source = "Config File"
			detail = fmt.Sprintf("%s key: %s", c.ConfigFile, o.Env)
		} else if envVal != "" {
			finalVal = envVal
			source = "Environment"
			detail = fmt.Sprintf("%s=%s", o.Env, envVal)
		} else if o.Default == 0 {
			// For IntOptions with default 0, they might not be explicitly set anywhere
			// but we still want to show them as Default 0
		}

		infos = append(infos, optionInfo{o.Name, finalVal, source, detail})
	}

	for _, o := range config.BoolOptions {
		finalVal := strconv.FormatBool(o.Default)
		source := "Default"
		detail := ""

		flagVal := ""
		if f := c.rootCmd.fs.Lookup(o.Name); f != nil {
			flagVal = f.Value.String()
		}

		envVal := os.Getenv(o.Env)
		fileVal := fileVals[o.Env]

		cliSet := setFlags[o.Name] && o.Name != ""

		var b bool
		if cliSet && flagVal != "" {
			b, _ = strconv.ParseBool(flagVal)
			finalVal = strconv.FormatBool(b)
			source = "Arg"
			detail = fmt.Sprintf("--%s", o.Name)
		} else if fileVal != "" {
			b, _ = strconv.ParseBool(fileVal)
			finalVal = strconv.FormatBool(b)
			source = "Config File"
			detail = fmt.Sprintf("%s key: %s", c.ConfigFile, o.Env)
		} else if envVal != "" {
			b, _ = strconv.ParseBool(envVal)
			finalVal = strconv.FormatBool(b)
			source = "Environment"
			detail = fmt.Sprintf("%s=%s", o.Env, envVal)
		}

		infos = append(infos, optionInfo{o.Name, finalVal, source, detail})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	for _, info := range infos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", info.Name, info.FinalVal, info.Source, info.Detail)
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
