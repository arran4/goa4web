package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

// notificationsSendCmd implements "notifications send".
type notificationsSendCmd struct {
	*notificationsCmd
	fs      *flag.FlagSet
	message string
	link    string
	role    string
	users   string
	jsonOut bool
}

func parseNotificationsSendCmd(parent *notificationsCmd, args []string) (*notificationsSendCmd, error) {
	c := &notificationsSendCmd{notificationsCmd: parent}
	fs, _, err := parseFlags("send", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.message, "message", "", "notification message")
		fs.StringVar(&c.link, "link", "", "link to include in the notification")
		fs.StringVar(&c.role, "role", "", "target role (defaults to all users)")
		fs.StringVar(&c.users, "users", "", "comma-separated usernames to notify")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.fs.Usage = c.Usage
	return c, nil
}

// Usage prints command usage information with examples.
func (c *notificationsSendCmd) Usage() {
	executeUsage(c.fs.Output(), "notifications_send_usage.txt", c)
}

func (c *notificationsSendCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*notificationsSendCmd)(nil)

func (c *notificationsSendCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	ids, err := c.resolveRecipientIDs(ctx, queries)
	if err != nil {
		return err
	}
	for _, id := range ids {
		err := queries.SystemCreateNotification(ctx, db.SystemCreateNotificationParams{
			RecipientID: id,
			Link:        sql.NullString{String: c.link, Valid: c.link != ""},
			Message:     sql.NullString{String: c.message, Valid: c.message != ""},
		})
		if err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}
	}
	if c.jsonOut {
		out := map[string]interface{}{
			"count":    len(ids),
			"user_ids": ids,
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintln(c.fs.Output(), string(b))
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Recipients\tStatus")
	fmt.Fprintf(w, "%d\tqueued\n", len(ids))
	return w.Flush()
}

func (c *notificationsSendCmd) resolveRecipientIDs(ctx context.Context, queries *db.Queries) ([]int32, error) {
	if c.users != "" {
		var ids []int32
		for _, name := range strings.Split(c.users, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: name, Valid: true})
			if err != nil {
				return nil, fmt.Errorf("get user %s: %w", name, err)
			}
			ids = append(ids, u.Idusers)
		}
		return ids, nil
	}
	if c.role != "" && c.role != "anyone" {
		rows, err := queries.AdminListUserIDsByRole(ctx, c.role)
		if err != nil {
			return nil, fmt.Errorf("list role: %w", err)
		}
		return rows, nil
	}
	rows, err := queries.AdminListAllUserIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return rows, nil
}
