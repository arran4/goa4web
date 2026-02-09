package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/sign"
)

// linksVerifyCmd implements "links verify".
type linksVerifyCmd struct {
	*linksCmd
	fs  *flag.FlagSet
	ts  string
	sig string
	url string
}

func parseLinksVerifyCmd(parent *linksCmd, args []string) (*linksVerifyCmd, error) {
	c := &linksVerifyCmd{linksCmd: parent}
	fs, rest, err := parseFlags("verify", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.ts, "ts", "", "timestamp")
		fs.StringVar(&c.sig, "sig", "", "signature")
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

func (c *linksVerifyCmd) Run() error {
	cfg := c.rootCmd.cfg
	key, err := config.LoadOrCreateLinkSignSecret(core.OSFS{}, cfg.LinkSignSecret, cfg.LinkSignSecretFile)
	if err != nil {
		return fmt.Errorf("link sign secret: %w", err)
	}
	tsInt, err := strconv.ParseInt(c.ts, 10, 64)
	if err == nil {
		err = sign.Verify(c.url, c.sig, key, sign.WithExpiry(time.Unix(tsInt, 0)))
	}
	if err == nil {
		fmt.Println("valid")
	} else {
		fmt.Println("invalid")
	}
	return nil
}

func (c *linksVerifyCmd) Usage() {
	executeUsage(c.fs.Output(), "links_verify_usage.txt", c)
}

func (c *linksVerifyCmd) FlagGroups() []flagGroup {
	return append(c.linksCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*linksVerifyCmd)(nil)
