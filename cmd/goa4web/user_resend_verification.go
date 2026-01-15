package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
)

type userResendVerificationCmd struct {
	*userCmd
	since    time.Duration
	allTime  bool
	dryRun   bool
	userID   int
	username string
}

func parseUserResendVerificationCmd(parent *userCmd, args []string) (*userResendVerificationCmd, error) {
	c := &userResendVerificationCmd{userCmd: parent}
	c.fs = newFlagSet("resend-verification")
	c.fs.DurationVar(&c.since, "since", 0, "Duration to look back for unverified emails (e.g., 24h).")
	c.fs.BoolVar(&c.allTime, "all-time", false, "Resend for all unverified emails, ignoring time filter.")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "List emails that would be processed without taking action.")
	c.fs.IntVar(&c.userID, "user-id", 0, "User ID to resend verification emails for.")
	c.fs.StringVar(&c.username, "username", "", "Username to resend verification emails for.")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userResendVerificationCmd) Run() error {
	if !c.allTime && c.since <= 0 && c.userID == 0 && c.username == "" {
		return fmt.Errorf("must specify -since duration, -all-time, or a user")
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

	if c.userID != 0 || c.username != "" {
		uid, err := resolveUserID(c.rootCmd.Context(), queries, c.userID, c.username)
		if err != nil {
			return err
		}
		all, err := queries.AdminListUserEmails(c.rootCmd.Context(), uid)
		if err != nil {
			return fmt.Errorf("list user emails: %w", err)
		}
		for _, e := range all {
			if !e.VerifiedAt.Valid {
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
	} else if c.allTime {
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
		cutoff := time.Now().Add(-c.since)
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

	if c.dryRun {
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
			page = cfg.HTTPHostname + path
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

func (c *userResendVerificationCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage: %s user resend-verification [flags]\n", os.Args[0])
	c.fs.PrintDefaults()
}

func (c *userResendVerificationCmd) loadConfig() (*config.RuntimeConfig, error) {
	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}
	return config.NewRuntimeConfig(
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	), nil
}
