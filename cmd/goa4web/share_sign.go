package main

import (
	"flag"
	"fmt"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/sign/signutil"
)

// shareSignCmd implements "share sign".
type shareSignCmd struct {
	*shareCmd
	fs       *flag.FlagSet
	Duration string
	NoExpiry bool
	url      string
}

func parseShareSignCmd(parent *shareCmd, args []string) (*shareSignCmd, error) {
	c := &shareSignCmd{shareCmd: parent}
	fs, rest, err := parseFlags("sign", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Duration, "duration", "24h", "validity duration, e.g. 24h")
		fs.BoolVar(&c.NoExpiry, "no-expiry", false, "generate link without expiry")
	})
	if err != nil {
		return nil, err
	}
	if len(rest) == 0 {
		return nil, fmt.Errorf("url argument required (use path without /shared/, e.g., /private/topic/2/thread/1)")
	}
	c.fs = fs
	c.url = rest[0]
	return c, nil
}

func (c *shareSignCmd) Run() error {
	cfg := c.rootCmd.cfg
	key, err := config.LoadOrCreateShareSignSecret(core.OSFS{}, cfg.ShareSignSecret, cfg.ShareSignSecretFile)
	if err != nil {
		return fmt.Errorf("share sign secret: %w", err)
	}
	signed, err := signutil.SignSharePath(c.url, key, c.Duration, c.NoExpiry)
	if err != nil {
		return fmt.Errorf("sign url: %w", err)
	}
	fmt.Println(signed)
	return nil
}

func (c *shareSignCmd) Usage() {
	executeUsage(c.fs.Output(), "share_sign_usage.txt", c)
}

func (c *shareSignCmd) FlagGroups() []flagGroup {
	return append(c.shareCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*shareSignCmd)(nil)
