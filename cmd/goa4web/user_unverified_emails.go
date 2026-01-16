package main

import (
	"database/sql"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
)

type userUnverifiedEmailsCmd struct {
	*userCmd
}

func parseUserUnverifiedEmailsCmd(parent *userCmd, args []string) (*userUnverifiedEmailsCmd, error) {
	c := &userUnverifiedEmailsCmd{userCmd: parent}
	c.fs = newFlagSet("unverified-emails")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userUnverifiedEmailsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.Usage()
		return fmt.Errorf("missing unverified-emails subcommand")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		return c.runList(args[1:])
	case "resend":
		return c.runResend(args[1:])
	case "expunge":
		return c.runExpunge(args[1:])
	default:
		c.Usage()
		return fmt.Errorf("unknown unverified-emails subcommand %q", args[0])
	}
}

func (c *userUnverifiedEmailsCmd) Usage() {
	executeUsage(c.fs.Output(), "user_unverified_emails_usage.txt", c)
}

func (c *userUnverifiedEmailsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userUnverifiedEmailsCmd)(nil)

func (c *userUnverifiedEmailsCmd) runList(args []string) error {
	fs := newFlagSet("list")
	userID := fs.Int("user-id", 0, "User ID to list unverified emails for (optional)")
	username := fs.String("username", "", "Username to list unverified emails for (optional)")
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

	var emails []*db.UserEmail
	if *userID != 0 || *username != "" {
		uid, err := resolveUserID(c.rootCmd.Context(), queries, *userID, *username)
		if err != nil {
			return err
		}
		// We need ListUnverifiedEmailsByUserID?
		// Currently we have AdminListUserEmails (all) and SystemListVerifiedEmailsByUserID.
		// AdminListUserEmails returns all, we can filter in Go.
		all, err := queries.AdminListUserEmails(c.rootCmd.Context(), uid)
		if err != nil {
			return fmt.Errorf("list user emails: %w", err)
		}
		for _, e := range all {
			if !e.VerifiedAt.Valid {
				emails = append(emails, e)
			}
		}
	} else {
		// List all unverified emails
		all, err := queries.SystemListAllUnverifiedEmails(c.rootCmd.Context())
		if err != nil {
			return fmt.Errorf("list all unverified emails: %w", err)
		}
		emails = all
	}

	w := tabwriter.NewWriter(c.fs.Output(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tUserID\tEmail\tExpires")
	for _, e := range emails {
		expires := "N/A"
		if e.VerificationExpiresAt.Valid {
			expires = e.VerificationExpiresAt.Time.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%d\t%d\t%s\t%s\n", e.ID, e.UserID, e.Email, expires)
	}
	w.Flush()
	return nil
}

func (c *userUnverifiedEmailsCmd) runResend(args []string) error {
	fs := newFlagSet("resend")
	since := fs.Duration("since", 0, "Duration to look back for unverified emails (e.g., 24h).")
	allTime := fs.Bool("all-time", false, "Resend for all unverified emails, ignoring time filter.")
	dryRun := fs.Bool("dry-run", false, "List emails that would be processed without taking action.")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if !*allTime && *since <= 0 {
		return fmt.Errorf("must specify -since duration or -all-time")
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
	notifier := notif.New(notif.WithQueries(queries), notif.WithConfig(cfg))

	// Common struct to hold results
	type EmailRow struct {
		ID                    int32
		UserID                int32
		Email                 string
		VerifiedAt            sql.NullTime
		LastVerificationCode  sql.NullString
		VerificationExpiresAt sql.NullTime
		NotificationPriority  int32
	}

	var rows []EmailRow

	if *allTime {
		es, err := queries.SystemListAllUnverifiedEmails(c.rootCmd.Context())
		if err != nil {
			return fmt.Errorf("list all unverified emails: %w", err)
		}
		for _, e := range es {
			rows = append(rows, EmailRow{
				ID:                    e.ID,
				UserID:                e.UserID,
				Email:                 e.Email,
				VerifiedAt:            e.VerifiedAt,
				LastVerificationCode:  e.LastVerificationCode,
				VerificationExpiresAt: e.VerificationExpiresAt,
				NotificationPriority:  e.NotificationPriority,
			})
		}
	} else {
		cutoff := time.Now().Add(-*since)
		es, err := queries.SystemListUnverifiedEmailsCreatedAfter(c.rootCmd.Context(), sql.NullTime{Time: cutoff, Valid: true})
		if err != nil {
			return fmt.Errorf("list unverified emails: %w", err)
		}
		for _, e := range es {
			rows = append(rows, EmailRow{
				ID:                    e.ID,
				UserID:                e.UserID,
				Email:                 e.Email,
				VerifiedAt:            e.VerifiedAt,
				LastVerificationCode:  e.LastVerificationCode,
				VerificationExpiresAt: e.VerificationExpiresAt,
				NotificationPriority:  e.NotificationPriority,
			})
		}
	}

	c.Infof("Found %d unverified emails to process", len(rows))

	if *dryRun {
		for _, ue := range rows {
			fmt.Fprintf(c.fs.Output(), "Would resend verification to: UserID=%d, Email=%s\n", ue.UserID, ue.Email)
		}
		return nil
	}

	expiryHours := cfg.EmailVerificationExpiryHours
	if expiryHours <= 0 {
		expiryHours = 24
	}

	count := 0
	for _, ue := range rows {
		// Logic duplicated from previous user_resend_verification.go
		// We could refactor to a shared helper but this is fine for now.
		code := generateVerificationCode()
		expire := time.Now().Add(time.Duration(expiryHours) * time.Hour)

		if err := queries.SystemUpdateVerificationCode(c.rootCmd.Context(), db.SystemUpdateVerificationCodeParams{
			LastVerificationCode:  sql.NullString{String: code, Valid: true},
			VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true},
			ID:                    ue.ID,
		}); err != nil {
			fmt.Fprintf(c.fs.Output(), "Failed to update verification code for email %s (ID %d): %v\n", ue.Email, ue.ID, err)
			continue
		}

		path := "/usr/email/verify?code=" + code
		page := "http://localhost" + path
		if cfg.HTTPHostname != "" {
			page = cfg.HTTPHostname + path // Ensure HTTPHostname has no trailing slash or handle it
		}

		user, err := queries.SystemGetUserByID(c.rootCmd.Context(), ue.UserID)
		username := ""
		if err == nil {
			username = user.Username.String
		}

		data := map[string]any{
			"page":             page,
			"email":            ue.Email,
			"URL":              page,
			"VerificationCode": code,
			"Token":            code,
			"Username":         username,
			"ExpiresAt":        expire,
		}

		et := notif.NewEmailTemplates("verifyEmail")
		msg, err := notifier.RenderEmailFromTemplates(c.rootCmd.Context(), ue.Email, et, data)
		if err != nil {
			fmt.Fprintf(c.fs.Output(), "Failed to render email for %s: %v\n", ue.Email, err)
			continue
		}

		if err := queries.InsertPendingEmail(c.rootCmd.Context(), db.InsertPendingEmailParams{
			ToUserID:    sql.NullInt32{Int32: ue.UserID, Valid: true},
			Body:        string(msg),
			DirectEmail: false,
		}); err != nil {
			fmt.Fprintf(c.fs.Output(), "Failed to queue email for %s: %v\n", ue.Email, err)
			continue
		}
		count++
	}
	c.Infof("Queued %d verification emails", count)
	return nil
}

func (c *userUnverifiedEmailsCmd) runExpunge(args []string) error {
	fs := newFlagSet("expunge")
	olderThan := fs.Duration("older-than", 0, "Duration to define 'older than' (e.g., 72h).")
	dryRun := fs.Bool("dry-run", false, "List emails that would be deleted without taking action.")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *olderThan <= 0 {
		return fmt.Errorf("missing or invalid -older-than duration")
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

	cutoff := time.Now().Add(-*olderThan)

	if *dryRun {
		es, err := queries.SystemListUnverifiedEmailsExpiresBefore(c.rootCmd.Context(), sql.NullTime{Time: cutoff, Valid: true})
		if err != nil {
			return fmt.Errorf("list unverified emails: %w", err)
		}
		for _, e := range es {
			fmt.Fprintf(c.fs.Output(), "Would expunge: ID=%d, UserID=%d, Email=%s, Expires=%v\n", e.ID, e.UserID, e.Email, e.VerificationExpiresAt.Time)
		}
		return nil
	}

	res, err := queries.SystemDeleteUnverifiedEmailsExpiresBefore(c.rootCmd.Context(), sql.NullTime{Time: cutoff, Valid: true})
	if err != nil {
		return fmt.Errorf("expunge unverified emails: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		c.Infof("Expunged unknown number of unverified emails")
	} else {
		c.Infof("Expunged %d unverified emails", rows)
	}
	return nil
}

func (c *userUnverifiedEmailsCmd) loadConfig() (*config.RuntimeConfig, error) {
	return c.rootCmd.RuntimeConfig()
}
