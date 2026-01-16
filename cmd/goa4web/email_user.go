package main

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
)

type emailUserCmd struct {
	*emailCmd
}

func parseEmailUserCmd(parent *emailCmd, args []string) (*emailUserCmd, error) {
	c := &emailUserCmd{emailCmd: parent}
	c.fs = newFlagSet("user")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailUserCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.Usage()
		return fmt.Errorf("missing user subcommand")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		return c.runList(args[1:])
	case "add":
		return c.runAdd(args[1:])
	case "delete":
		return c.runDelete(args[1:])
	case "update":
		return c.runUpdate(args[1:])
	default:
		c.Usage()
		return fmt.Errorf("unknown user subcommand %q", args[0])
	}
}

func (c *emailUserCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage: %s email user [subcommand] [flags]\n\nSubcommands: list, add, delete, update\n", os.Args[0])
	c.fs.PrintDefaults()
}

func (c *emailUserCmd) runList(args []string) error {
	fs := newFlagSet("list")
	userID := fs.Int("user-id", 0, "User ID to list verified emails for")
	username := fs.String("username", "", "Username to list verified emails for")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := c.loadConfig()
	if err != nil {
		return err
	}
	d, err := c.rootCmd.InitDB(cfg)
	if err != nil {
		return err
	}
	defer d.Close()
	queries := db.New(d)

	uid, err := resolveUserID(c.rootCmd.Context(), queries, *userID, *username)
	if err != nil {
		return err
	}

	emails, err := queries.SystemListVerifiedEmailsByUserID(c.rootCmd.Context(), uid)
	if err != nil {
		return fmt.Errorf("list verified emails: %w", err)
	}

	w := tabwriter.NewWriter(c.fs.Output(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tUserID\tEmail\tVerified\tPriority")
	for _, e := range emails {
		verified := "No"
		if e.VerifiedAt.Valid {
			verified = e.VerifiedAt.Time.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%d\n", e.ID, e.UserID, e.Email, verified, e.NotificationPriority)
	}
	w.Flush()
	return nil
}

func (c *emailUserCmd) runAdd(args []string) error {
	fs := newFlagSet("add")
	userID := fs.Int("user-id", 0, "User ID to add email to")
	username := fs.String("username", "", "Username to add email to")
	email := fs.String("email", "", "Email address to add")
	priority := fs.Int("priority", 0, "Notification priority")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *email == "" {
		return fmt.Errorf("missing -email")
	}

	cfg, err := c.loadConfig()
	if err != nil {
		return err
	}
	d, err := c.rootCmd.InitDB(cfg)
	if err != nil {
		return err
	}
	defer d.Close()
	queries := db.New(d)

	uid, err := resolveUserID(c.rootCmd.Context(), queries, *userID, *username)
	if err != nil {
		return err
	}

	verifiedAt := sql.NullTime{Time: time.Now(), Valid: true}

	if err := queries.AdminAddUserEmail(c.rootCmd.Context(), db.AdminAddUserEmailParams{
		UserID:               uid,
		Email:                *email,
		VerifiedAt:           verifiedAt,
		NotificationPriority: int32(*priority),
	}); err != nil {
		return fmt.Errorf("add verified email: %w", err)
	}
	c.Infof("Verified email added successfully")
	return nil
}

func (c *emailUserCmd) runDelete(args []string) error {
	fs := newFlagSet("delete")
	id := fs.Int("id", 0, "Email ID to delete (use list to find ID)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == 0 {
		return fmt.Errorf("missing -id")
	}

	cfg, err := c.loadConfig()
	if err != nil {
		return err
	}
	d, err := c.rootCmd.InitDB(cfg)
	if err != nil {
		return err
	}
	defer d.Close()
	queries := db.New(d)

	if err := queries.AdminDeleteUserEmail(c.rootCmd.Context(), int32(*id)); err != nil {
		return fmt.Errorf("delete email: %w", err)
	}
	c.Infof("Email deleted successfully")
	return nil
}

func (c *emailUserCmd) runUpdate(args []string) error {
	fs := newFlagSet("update")
	id := fs.Int("id", 0, "Email ID to update")
	email := fs.String("email", "", "New email address (optional)")
	verified := fs.String("verified", "", "Set verification status (true/false/now) (optional)")
	priority := fs.Int("priority", -1, "New notification priority (optional)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == 0 {
		return fmt.Errorf("missing -id")
	}

	cfg, err := c.loadConfig()
	if err != nil {
		return err
	}
	d, err := c.rootCmd.InitDB(cfg)
	if err != nil {
		return err
	}
	defer d.Close()
	queries := db.New(d)

	current, err := queries.AdminGetUserEmailByID(c.rootCmd.Context(), int32(*id))
	if err != nil {
		return fmt.Errorf("get email: %w", err)
	}

	newEmail := current.Email
	if *email != "" {
		newEmail = *email
	}
	newPriority := current.NotificationPriority
	if *priority != -1 {
		newPriority = int32(*priority)
	}
	newVerified := current.VerifiedAt
	if *verified != "" {
		switch *verified {
		case "true", "now":
			newVerified = sql.NullTime{Time: time.Now(), Valid: true}
		case "false":
			newVerified = sql.NullTime{Valid: false}
		}
	}

	if err := queries.AdminUpdateUserEmailDetails(c.rootCmd.Context(), db.AdminUpdateUserEmailDetailsParams{
		ID:                   int32(*id),
		Email:                newEmail,
		VerifiedAt:           newVerified,
		NotificationPriority: newPriority,
	}); err != nil {
		return fmt.Errorf("update email: %w", err)
	}
	c.Infof("Email updated successfully")
	return nil
}

func (c *emailUserCmd) loadConfig() (*config.RuntimeConfig, error) {
	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}
	return config.NewRuntimeConfig(
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	), nil
}
