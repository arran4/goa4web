package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

type pageSizeCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parsePageSizeCmd(parent *rootCmd, args []string) (*pageSizeCmd, error) {
	c := &pageSizeCmd{rootCmd: parent}
	c.fs = newFlagSet("page-size")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *pageSizeCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing page-size command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "show":
		cmd, err := parsePageSizeShowCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("show: %w", err)
		}
		return cmd.Run()
	case "set":
		cmd, err := parsePageSizeSetCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("set: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown page-size command %q", args[0])
	}
}

func (c *pageSizeCmd) Usage() {
	executeUsage(c.fs.Output(), "page_size_usage.txt", c)
}

func (c *pageSizeCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

type pageSizeShowCmd struct {
	*pageSizeCmd
	fs *flag.FlagSet
}

func parsePageSizeShowCmd(parent *pageSizeCmd, args []string) (*pageSizeShowCmd, error) {
	c := &pageSizeShowCmd{pageSizeCmd: parent}
	c.fs = newFlagSet("show")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *pageSizeShowCmd) Run() error {
	return writePageSizeJSON(c.rootCmd.cfg)
}

type pageSizeSetCmd struct {
	*pageSizeCmd
	fs  *flag.FlagSet
	min int
	max int
	def int
}

func parsePageSizeSetCmd(parent *pageSizeCmd, args []string) (*pageSizeSetCmd, error) {
	c := &pageSizeSetCmd{pageSizeCmd: parent}
	c.fs = newFlagSet("set")
	c.fs.IntVar(&c.min, "min", 0, "minimum page size (0 keeps current)")
	c.fs.IntVar(&c.max, "max", 0, "maximum page size (0 keeps current)")
	c.fs.IntVar(&c.def, "default", 0, "default page size (0 keeps current)")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *pageSizeSetCmd) Run() error {
	if c.min == 0 && c.max == 0 && c.def == 0 {
		return fmt.Errorf("at least one of -min, -max, or -default is required")
	}
	if c.min < 0 || c.max < 0 || c.def < 0 {
		return fmt.Errorf("page size values must be positive")
	}
	cfg := c.rootCmd.cfg
	newMin := cfg.PageSizeMin
	newMax := cfg.PageSizeMax
	newDef := cfg.PageSizeDefault
	if c.min != 0 {
		newMin = c.min
	}
	if c.max != 0 {
		newMax = c.max
	}
	if c.def != 0 {
		newDef = c.def
	}
	if newMin > newMax {
		return fmt.Errorf("min page size %d cannot exceed max page size %d", newMin, newMax)
	}
	if newDef < newMin || newDef > newMax {
		return fmt.Errorf("default page size %d must be between %d and %d", newDef, newMin, newMax)
	}

	config.UpdatePaginationConfig(cfg, c.min, c.max, c.def)

	if c.min != 0 {
		if err := config.UpdateConfigKey(core.OSFS{}, c.rootCmd.ConfigFile, config.EnvPageSizeMin, strconv.Itoa(cfg.PageSizeMin)); err != nil {
			return fmt.Errorf("update config min: %w", err)
		}
	}
	if c.max != 0 {
		if err := config.UpdateConfigKey(core.OSFS{}, c.rootCmd.ConfigFile, config.EnvPageSizeMax, strconv.Itoa(cfg.PageSizeMax)); err != nil {
			return fmt.Errorf("update config max: %w", err)
		}
	}
	if c.def != 0 {
		if err := config.UpdateConfigKey(core.OSFS{}, c.rootCmd.ConfigFile, config.EnvPageSizeDefault, strconv.Itoa(cfg.PageSizeDefault)); err != nil {
			return fmt.Errorf("update config default: %w", err)
		}
	}

	return writePageSizeJSON(cfg)
}

type pageSizeOutput struct {
	Min     int `json:"min"`
	Max     int `json:"max"`
	Default int `json:"default"`
}

func writePageSizeJSON(cfg *config.RuntimeConfig) error {
	out := pageSizeOutput{
		Min:     cfg.PageSizeMin,
		Max:     cfg.PageSizeMax,
		Default: cfg.PageSizeDefault,
	}
	data, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("marshal page size: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

var _ usageData = (*pageSizeCmd)(nil)
