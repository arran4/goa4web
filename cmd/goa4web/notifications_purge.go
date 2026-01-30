package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// notificationsPurgeReadCmd implements "notifications purge-read".
type notificationsPurgeReadCmd struct {
	*notificationsCmd
	fs      *flag.FlagSet
	userID  int
	jsonOut bool
}

func parseNotificationsPurgeReadCmd(parent *notificationsCmd, args []string) (*notificationsPurgeReadCmd, error) {
	c := &notificationsPurgeReadCmd{notificationsCmd: parent}
	fs, _, err := parseFlags("purge-read", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.userID, "user", 0, "only purge notifications for this user ID")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

// Usage prints command usage information with examples.
func (c *notificationsPurgeReadCmd) Usage() {
	executeUsage(c.fs.Output(), "notifications_purge_read_usage.txt", c)
}

func (c *notificationsPurgeReadCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*notificationsPurgeReadCmd)(nil)

func (c *notificationsPurgeReadCmd) Run() error {
	if len(c.fs.Args()) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(c.fs.Args(), " "))
	}
	if c.userID < 0 {
		return fmt.Errorf("user must be a positive ID")
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)

	output := notificationsPurgeOutput{
		Action: "purge-read",
	}
	if c.userID > 0 {
		userID := int32(c.userID)
		output.UserID = &userID
		purged, errs, err := purgeReadNotificationsForUser(ctx, queries, userID)
		if err != nil {
			output.Errors = append(output.Errors, notificationsPurgeError{Message: err.Error()})
			c.renderNotificationsPurgeOutput(output)
			return err
		}
		output.Purged = purged
		output.Errors = append(output.Errors, errs...)
		c.renderNotificationsPurgeOutput(output)
		if len(output.Errors) > 0 {
			return fmt.Errorf("purge-read completed with %d errors", len(output.Errors))
		}
		return nil
	}

	purged, err := purgeReadNotifications(ctx, conn)
	if err != nil {
		output.Errors = append(output.Errors, notificationsPurgeError{Message: err.Error()})
		c.renderNotificationsPurgeOutput(output)
		return err
	}
	output.Purged = purged
	c.renderNotificationsPurgeOutput(output)
	if len(output.Errors) > 0 {
		return fmt.Errorf("purge-read completed with %d errors", len(output.Errors))
	}
	return nil
}

// notificationsPurgeSelectedCmd implements "notifications purge-selected".
type notificationsPurgeSelectedCmd struct {
	*notificationsCmd
	fs      *flag.FlagSet
	userID  int
	jsonOut bool
}

func parseNotificationsPurgeSelectedCmd(parent *notificationsCmd, args []string) (*notificationsPurgeSelectedCmd, error) {
	c := &notificationsPurgeSelectedCmd{notificationsCmd: parent}
	fs, _, err := parseFlags("purge-selected", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.userID, "user", 0, "only purge notifications for this user ID")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

// Usage prints command usage information with examples.
func (c *notificationsPurgeSelectedCmd) Usage() {
	executeUsage(c.fs.Output(), "notifications_purge_selected_usage.txt", c)
}

func (c *notificationsPurgeSelectedCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*notificationsPurgeSelectedCmd)(nil)

func (c *notificationsPurgeSelectedCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing notification IDs")
	}
	if c.userID < 0 {
		return fmt.Errorf("user must be a positive ID")
	}

	ids := make([]int32, 0, len(args))
	for _, arg := range args {
		id, err := parsePositiveInt32(arg)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	queries, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()

	output := notificationsPurgeOutput{
		Action: "purge-selected",
		IDs:    ids,
	}
	if c.userID > 0 {
		userID := int32(c.userID)
		output.UserID = &userID
	}

	purged, errs := purgeSelectedNotifications(ctx, queries, ids, output.UserID)
	output.Purged = purged
	output.Errors = append(output.Errors, errs...)
	c.renderNotificationsPurgeOutput(output)
	if len(output.Errors) > 0 {
		return fmt.Errorf("purge-selected completed with %d errors", len(output.Errors))
	}
	return nil
}

type notificationsPurgeError struct {
	ID      int32  `json:"id,omitempty"`
	Message string `json:"message"`
}

type notificationsPurgeOutput struct {
	Action string                    `json:"action"`
	UserID *int32                    `json:"user_id,omitempty"`
	IDs    []int32                   `json:"ids,omitempty"`
	Purged int                       `json:"purged"`
	Errors []notificationsPurgeError `json:"errors,omitempty"`
}

func (c *notificationsPurgeReadCmd) renderNotificationsPurgeOutput(output notificationsPurgeOutput) {
	renderNotificationsPurgeOutput(c.fs.Output(), output, c.jsonOut)
}

func (c *notificationsPurgeSelectedCmd) renderNotificationsPurgeOutput(output notificationsPurgeOutput) {
	renderNotificationsPurgeOutput(c.fs.Output(), output, c.jsonOut)
}

func renderNotificationsPurgeOutput(outWriter io.Writer, output notificationsPurgeOutput, jsonOut bool) {
	if jsonOut {
		b, _ := json.MarshalIndent(output, "", "  ")
		fmt.Fprintln(outWriter, string(b))
		return
	}

	header := fmt.Sprintf("Purged %d notifications", output.Purged)
	if output.UserID != nil {
		header = fmt.Sprintf("Purged %d notifications for user %d", output.Purged, *output.UserID)
	}
	fmt.Fprintln(outWriter, header+".")
	if len(output.Errors) == 0 {
		return
	}
	fmt.Fprintf(outWriter, "Errors (%d):\n", len(output.Errors))
	for _, entry := range output.Errors {
		if entry.ID > 0 {
			fmt.Fprintf(outWriter, "- ID %d: %s\n", entry.ID, entry.Message)
			continue
		}
		fmt.Fprintf(outWriter, "- %s\n", entry.Message)
	}
}

func purgeReadNotifications(ctx context.Context, conn *sql.DB) (int, error) {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	queries := db.New(tx)
	if err := queries.AdminPurgeReadNotifications(ctx); err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("purge read notifications: %w", err)
	}

	var count int
	row := tx.QueryRowContext(ctx, "SELECT ROW_COUNT()")
	if err := row.Scan(&count); err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("read purge count: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit purge: %w", err)
	}
	return count, nil
}

func purgeReadNotificationsForUser(ctx context.Context, queries *db.Queries, userID int32) (int, []notificationsPurgeError, error) {
	// notificationsPurgePageSize is the batch size for paginating notifications.
	const notificationsPurgePageSize = int32(200)

	notifications, err := listNotificationsForUser(ctx, queries, userID, notificationsPurgePageSize)
	if err != nil {
		return 0, nil, fmt.Errorf("list notifications: %w", err)
	}
	cutoff := time.Now().Add(-24 * time.Hour)
	ids := make([]int32, 0, len(notifications))
	for _, notification := range notifications {
		if !notification.ReadAt.Valid {
			continue
		}
		if notification.ReadAt.Time.Before(cutoff) {
			ids = append(ids, notification.ID)
		}
	}

	purged, errs := purgeSelectedNotifications(ctx, queries, ids, &userID)
	return purged, errs, nil
}

func listNotificationsForUser(ctx context.Context, queries *db.Queries, userID int32, pageSize int32) ([]*db.Notification, error) {
	var out []*db.Notification
	var offset int32
	for {
		rows, err := queries.ListNotificationsForLister(ctx, db.ListNotificationsForListerParams{
			ListerID: userID,
			Limit:    pageSize,
			Offset:   offset,
		})
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return out, nil
		}
		out = append(out, rows...)
		if len(rows) < int(pageSize) {
			return out, nil
		}
		offset += pageSize
	}
}

func purgeSelectedNotifications(ctx context.Context, queries db.Querier, ids []int32, userID *int32) (int, []notificationsPurgeError) {
	purged := 0
	var errs []notificationsPurgeError
	for _, id := range ids {
		if id <= 0 {
			errs = append(errs, notificationsPurgeError{ID: id, Message: "invalid notification ID"})
			continue
		}
		n, err := queries.AdminGetNotification(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				errs = append(errs, notificationsPurgeError{ID: id, Message: "notification not found"})
				continue
			}
			errs = append(errs, notificationsPurgeError{ID: id, Message: fmt.Sprintf("lookup notification: %v", err)})
			continue
		}
		if userID != nil && n.UsersIdusers != *userID {
			errs = append(errs, notificationsPurgeError{ID: id, Message: fmt.Sprintf("notification belongs to user %d", n.UsersIdusers)})
			continue
		}
		if err := queries.AdminDeleteNotification(ctx, id); err != nil {
			errs = append(errs, notificationsPurgeError{ID: id, Message: fmt.Sprintf("delete notification: %v", err)})
			continue
		}
		purged++
	}
	return purged, errs
}

func parsePositiveInt32(value string) (int32, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid notification ID %q", value)
	}
	return int32(id), nil
}
