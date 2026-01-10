package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	linksign "github.com/arran4/goa4web/internal/linksign"
)

// linksSignCmd implements "links sign".
type linksSignCmd struct {
	*linksCmd
	fs       *flag.FlagSet
	Duration string
	NoExpiry bool
	url      string
}

func parseLinksSignCmd(parent *linksCmd, args []string) (*linksSignCmd, error) {
	c := &linksSignCmd{linksCmd: parent}
	fs, rest, err := parseFlags("sign", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Duration, "duration", "24h", "validity duration, e.g. 24h")
		fs.BoolVar(&c.NoExpiry, "no-expiry", false, "generate link without expiry")
	})
	if err != nil {
		return nil, err
	}
	if len(rest) == 0 {
		return nil, fmt.Errorf("url argument required")
	}
	c.fs = fs
	c.url = rest[0]
	return c, nil
}

func (c *linksSignCmd) Run() error {
	cfg := c.rootCmd.cfg
	key, err := config.LoadOrCreateLinkSignSecret(core.OSFS{}, cfg.LinkSignSecret, cfg.LinkSignSecretFile)
	if err != nil {
		return fmt.Errorf("link sign secret: %w", err)
	}
	linkSignExpiry, err := time.ParseDuration(cfg.LinkSignExpiry)
	if err != nil {
		return fmt.Errorf("parsing link sign expiry: %w", err)
	}
	signer := linksign.NewSigner(cfg, key, linkSignExpiry)
	var exp time.Time
	if !c.NoExpiry {
		d, err := time.ParseDuration(c.Duration)
		if err != nil {
			return fmt.Errorf("parse duration: %w", err)
		}
		exp = time.Now().Add(d)
	}
	signed := signer.SignedURL(c.url, exp)
	fmt.Println(signed)
	return nil
}

func (c *linksSignCmd) Usage() {
	executeUsage(c.fs.Output(), "links_sign_usage.txt", c)
}

func (c *linksSignCmd) FlagGroups() []flagGroup {
	return append(c.linksCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*linksSignCmd)(nil)
