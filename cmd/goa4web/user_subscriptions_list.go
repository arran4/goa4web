package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// defaultSubscriptionsFormat sets the default output format for subscription listings.
const defaultSubscriptionsFormat = "csv"

// userSubscriptionsListCmd implements "user subscriptions list".
type userSubscriptionsListCmd struct {
	*userSubscriptionsCmd
	fs            *flag.FlagSet
	userID        int
	username      string
	method        string
	patternPrefix string
	format        string
}

type subscriptionListOutput struct {
	UserID         int32  `json:"user_id"`
	SubscriptionID int32  `json:"subscription_id"`
	Pattern        string `json:"pattern"`
	Method         string `json:"method"`
}

func parseUserSubscriptionsListCmd(parent *userSubscriptionsCmd, args []string) (*userSubscriptionsListCmd, error) {
	c := &userSubscriptionsListCmd{userSubscriptionsCmd: parent}
	fs := newFlagSet("list")
	fs.IntVar(&c.userID, "user-id", 0, "User ID to list subscriptions for")
	fs.StringVar(&c.username, "username", "", "Username to list subscriptions for")
	fs.StringVar(&c.method, "method", "", "Filter by subscription method (email/internal)")
	fs.StringVar(&c.patternPrefix, "pattern-prefix", "", "Filter by pattern prefix")
	fs.StringVar(&c.format, "format", defaultSubscriptionsFormat, "Output format (csv/json)")
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userSubscriptionsListCmd) Run() error {
	if c.userID == 0 && c.username == "" {
		return fmt.Errorf("missing -user-id or -username")
	}
	method := strings.ToLower(strings.TrimSpace(c.method))
	if method != "" && method != "email" && method != "internal" {
		return fmt.Errorf("unsupported method %q", c.method)
	}
	format := strings.ToLower(strings.TrimSpace(c.format))
	if format == "" {
		format = defaultSubscriptionsFormat
	}
	if format != "csv" && format != "json" {
		return fmt.Errorf("unsupported format %q", c.format)
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := c.rootCmd.Context()
	queries := db.New(conn)
	uid, err := resolveUserID(ctx, queries, c.userID, c.username)
	if err != nil {
		return err
	}
	rows, err := queries.ListSubscriptionsByUser(ctx, uid)
	if err != nil {
		return fmt.Errorf("list subscriptions: %w", err)
	}

	filtered := make([]subscriptionListOutput, 0, len(rows))
	for _, row := range rows {
		if method != "" && strings.ToLower(row.Method) != method {
			continue
		}
		if c.patternPrefix != "" && !strings.HasPrefix(row.Pattern, c.patternPrefix) {
			continue
		}
		filtered = append(filtered, subscriptionListOutput{
			UserID:         uid,
			SubscriptionID: row.ID,
			Pattern:        row.Pattern,
			Method:         row.Method,
		})
	}

	if format == "json" {
		b, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		fmt.Fprintln(c.fs.Output(), string(b))
		return nil
	}

	writer := csv.NewWriter(c.fs.Output())
	if err := writer.Write([]string{"user_id", "subscription_id", "pattern", "method"}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}
	for _, row := range filtered {
		if err := writer.Write([]string{
			fmt.Sprintf("%d", row.UserID),
			fmt.Sprintf("%d", row.SubscriptionID),
			row.Pattern,
			row.Method,
		}); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("write csv: %w", err)
	}
	return nil
}

func (c *userSubscriptionsListCmd) Usage() {
	executeUsage(c.fs.Output(), "user_subscriptions_list_usage.txt", c)
}

func (c *userSubscriptionsListCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userSubscriptionsListCmd)(nil)
