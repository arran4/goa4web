package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
)

type userEmailCmd struct {
	*userCmd
}

type userEmailSubcmdUsage struct {
	*userEmailCmd
	fs *flag.FlagSet
}

func (u *userEmailSubcmdUsage) FlagGroups() []flagGroup {
	return []flagGroup{flagGroupFromFlagSet(u.fs)}
}

var _ usageData = (*userEmailSubcmdUsage)(nil)

func parseUserEmailCmd(parent *userCmd, args []string) (*userEmailCmd, error) {
	c := &userEmailCmd{userCmd: parent}
	c.fs = newFlagSet("email")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userEmailCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.Usage()
		return fmt.Errorf("missing email subcommand")
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
	case "verify":
		return c.runVerify(args[1:])
	case "unverify":
		return c.runUnverify(args[1:])
	default:
		c.Usage()
		return fmt.Errorf("unknown email subcommand %q", args[0])
	}
}

func (c *userEmailCmd) Usage() {
	executeUsage(c.fs.Output(), "user_email_usage.txt", c)
}

func (c *userEmailCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userEmailCmd)(nil)

func (c *userEmailCmd) runList(args []string) error {
	fs := newFlagSet("list")
	usage := &userEmailSubcmdUsage{userEmailCmd: c, fs: fs}
	fs.Usage = func() {
		_ = executeUsage(fs.Output(), "user_email_list_usage.txt", usage)
	}
	userID := fs.Int("user-id", 0, "User ID to list emails for")
	username := fs.String("username", "", "Username to list emails for")
	includeUnverified := fs.Bool("include-unverified", false, "Include unverified emails in results")
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

	var emails []*db.UserEmail
	if *includeUnverified {
		emails, err = queries.AdminListUserEmails(c.rootCmd.Context(), uid)
		if err != nil {
			return fmt.Errorf("list user emails: %w", err)
		}
	} else {
		emails, err = queries.SystemListVerifiedEmailsByUserID(c.rootCmd.Context(), uid)
		if err != nil {
			return fmt.Errorf("list verified emails: %w", err)
		}
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

func (c *userEmailCmd) runAdd(args []string) error {
	fs := newFlagSet("add")
	usage := &userEmailSubcmdUsage{userEmailCmd: c, fs: fs}
	fs.Usage = func() {
		_ = executeUsage(fs.Output(), "user_email_add_usage.txt", usage)
	}
	userID := fs.Int("user-id", 0, "User ID to add email to")
	username := fs.String("username", "", "Username to add email to")
	email := fs.String("email", "", "Email address to add")
	priority := fs.Int("priority", 0, "Notification priority")
	verified := fs.String("verified", "true", "Set verification status (true/false/now)")
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

	var verifiedAt sql.NullTime
	switch *verified {
	case "true", "now":
		verifiedAt = sql.NullTime{Time: time.Now(), Valid: true}
	case "false":
		verifiedAt = sql.NullTime{Valid: false}
	default:
		return fmt.Errorf("invalid -verified value %q", *verified)
	}

	if err := queries.AdminAddUserEmail(c.rootCmd.Context(), db.AdminAddUserEmailParams{
		UserID:               uid,
		Email:                *email,
		VerifiedAt:           verifiedAt,
		NotificationPriority: int32(*priority),
	}); err != nil {
		return fmt.Errorf("add verified email: %w", err)
	}
	c.Infof("Email added successfully")
	return nil
}

func (c *userEmailCmd) runDelete(args []string) error {
	fs := newFlagSet("delete")
	usage := &userEmailSubcmdUsage{userEmailCmd: c, fs: fs}
	fs.Usage = func() {
		_ = executeUsage(fs.Output(), "user_email_delete_usage.txt", usage)
	}
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

func (c *userEmailCmd) runUpdate(args []string) error {
	fs := newFlagSet("update")
	usage := &userEmailSubcmdUsage{userEmailCmd: c, fs: fs}
	fs.Usage = func() {
		_ = executeUsage(fs.Output(), "user_email_update_usage.txt", usage)
	}
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
		default:
			return fmt.Errorf("invalid -verified value %q", *verified)
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

func (c *userEmailCmd) runVerify(args []string) error {
	fs := newFlagSet("verify")
	usage := &userEmailSubcmdUsage{userEmailCmd: c, fs: fs}
	fs.Usage = func() {
		_ = executeUsage(fs.Output(), "user_email_verify_usage.txt", usage)
	}
	id := fs.Int("id", 0, "Email ID to verify")
	at := fs.String("at", "", "Verification timestamp (RFC3339, optional)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == 0 {
		return fmt.Errorf("missing -id")
	}

	verifiedAt := sql.NullTime{Time: time.Now(), Valid: true}
	if *at != "" {
		parsed, err := time.Parse(time.RFC3339, *at)
		if err != nil {
			return fmt.Errorf("parse -at: %w", err)
		}
		verifiedAt = sql.NullTime{Time: parsed, Valid: true}
	}

	return c.updateEmailVerification(int32(*id), verifiedAt, "Email verified successfully")
}

func (c *userEmailCmd) runUnverify(args []string) error {
	fs := newFlagSet("unverify")
	usage := &userEmailSubcmdUsage{userEmailCmd: c, fs: fs}
	fs.Usage = func() {
		_ = executeUsage(fs.Output(), "user_email_unverify_usage.txt", usage)
	}
	id := fs.Int("id", 0, "Email ID to unverify")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == 0 {
		return fmt.Errorf("missing -id")
	}

	return c.updateEmailVerification(int32(*id), sql.NullTime{Valid: false}, "Email unverified successfully")
}

func (c *userEmailCmd) updateEmailVerification(id int32, verifiedAt sql.NullTime, successMsg string) error {
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

	current, err := queries.AdminGetUserEmailByID(c.rootCmd.Context(), id)
	if err != nil {
		return fmt.Errorf("get email: %w", err)
	}

	if err := queries.AdminUpdateUserEmailDetails(c.rootCmd.Context(), db.AdminUpdateUserEmailDetailsParams{
		ID:                   id,
		Email:                current.Email,
		VerifiedAt:           verifiedAt,
		NotificationPriority: current.NotificationPriority,
	}); err != nil {
		return fmt.Errorf("update email verification: %w", err)
	}
	c.Infof("%s", successMsg)
	return nil
}

func (c *userEmailCmd) loadConfig() (*config.RuntimeConfig, error) {
	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}
	return config.NewRuntimeConfig(
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	), nil
}
