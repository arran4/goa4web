package main

import (
	"flag"
	"fmt"
	"net/mail"

	"github.com/arran4/goa4web/internal/db"
)

type userPasswordSendResetEmailCmd struct {
	*userCmd
	fs       *flag.FlagSet
	username string
	userID   int
}

func parseUserPasswordSendResetEmailCmd(parent *userCmd, args []string) (*userPasswordSendResetEmailCmd, error) {
	c := &userPasswordSendResetEmailCmd{userCmd: parent}
	fs := flag.NewFlagSet("send-reset-email", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.username, "username", "", "Username")
	fs.IntVar(&c.userID, "user-id", 0, "User ID")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userPasswordSendResetEmailCmd) Run() error {
	ctx := c.rootCmd.Context()

	d, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}
	defer d.Close()

	queries := db.New(d)
	cfg := c.rootCmd.cfg

	signedURL, uName, uid, err := getResetURL(ctx, queries, cfg, c.userID, c.username)
	if err != nil {
		return err
	}

	u, err := queries.SystemGetUserByID(ctx, uid)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if !u.Email.Valid || u.Email.String == "" {
		return fmt.Errorf("user has no email")
	}

	body := fmt.Sprintf("Hi %s,\n\nYou requested a password reset. Click the link below to set a new password:\n\n%s\n\nThis link is valid for 24 hours.", uName, signedURL)

	p, err := c.rootCmd.emailReg.ProviderFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("get email provider: %w", err)
	}
	if p == nil {
		return fmt.Errorf("email provider not configured")
	}

	to := mail.Address{Name: uName, Address: u.Email.String}
	msg := []byte(fmt.Sprintf("Subject: Password Reset\r\n\r\n%s", body))

	if err := p.Send(ctx, to, msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	fmt.Printf("Reset email sent to %s (%s)\n", uName, u.Email.String)
	return nil
}

// Usage prints command usage information with examples.
func (c *userPasswordSendResetEmailCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage of %s:\n", c.fs.Name())
	c.fs.PrintDefaults()
}

func (c *userPasswordSendResetEmailCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userPasswordSendResetEmailCmd)(nil)
