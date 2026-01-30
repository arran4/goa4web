package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// notificationsListCmd implements "notifications list".
type notificationsListCmd struct {
	*notificationsCmd
	fs       *flag.FlagSet
	userID   int
	unread   bool
	read     bool
	category string
	limit    int
	offset   int
	jsonOut  bool
}

func parseNotificationsListCmd(parent *notificationsCmd, args []string) (*notificationsListCmd, error) {
	c := &notificationsListCmd{notificationsCmd: parent}
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.userID, "user", 0, "filter by user ID")
		fs.BoolVar(&c.unread, "unread", false, "only unread notifications")
		fs.BoolVar(&c.read, "read", false, "only read notifications")
		fs.StringVar(&c.category, "category", "", "filter by link category (first path segment)")
		fs.IntVar(&c.limit, "limit", 50, "limit number of notifications")
		fs.IntVar(&c.offset, "offset", 0, "offset into the notification list")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

// Usage prints command usage information with examples.
func (c *notificationsListCmd) Usage() {
	executeUsage(c.fs.Output(), "notifications_list_usage.txt", c)
}

func (c *notificationsListCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*notificationsListCmd)(nil)

func (c *notificationsListCmd) Run() error {
	if len(c.fs.Args()) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(c.fs.Args(), " "))
	}
	if c.unread && c.read {
		return fmt.Errorf("cannot combine --read and --unread")
	}
	if c.userID < 0 {
		return fmt.Errorf("user must be a positive ID")
	}
	if c.limit <= 0 {
		return fmt.Errorf("limit must be greater than zero")
	}
	if c.offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	limit := int32(c.limit)
	offset := int32(c.offset)

	var rows []*db.Notification
	if c.userID > 0 {
		userID := int32(c.userID)
		if c.unread {
			rows, err = queries.ListUnreadNotificationsForLister(ctx, db.ListUnreadNotificationsForListerParams{
				ListerID: userID,
				Limit:    limit,
				Offset:   offset,
			})
		} else {
			rows, err = queries.ListNotificationsForLister(ctx, db.ListNotificationsForListerParams{
				ListerID: userID,
				Limit:    limit,
				Offset:   offset,
			})
		}
		if err != nil {
			return fmt.Errorf("list notifications: %w", err)
		}
	} else {
		queryLimit := limit + offset
		rows, err = queries.AdminListRecentNotifications(ctx, queryLimit)
		if err != nil {
			return fmt.Errorf("list notifications: %w", err)
		}
		if offset > 0 && int(offset) < len(rows) {
			rows = rows[offset:]
		} else if offset > 0 {
			rows = rows[:0]
		}
	}

	filtered := rows[:0]
	for _, n := range rows {
		if c.read && !n.ReadAt.Valid {
			continue
		}
		if c.unread && n.ReadAt.Valid {
			continue
		}
		if c.category != "" {
			category := notificationCategory(n.Link)
			if !strings.EqualFold(category, c.category) {
				continue
			}
		}
		filtered = append(filtered, n)
	}

	if c.jsonOut {
		out := make([]notificationListItem, 0, len(filtered))
		for _, n := range filtered {
			out = append(out, newNotificationListItem(n))
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tUserID\tUnread\tCategory\tCreated\tRead\tMessage\tLink")
	for _, n := range filtered {
		item := newNotificationListItem(n)
		readAt := "-"
		if item.ReadAt != nil {
			readAt = *item.ReadAt
		}
		message := ""
		if item.Message != nil {
			message = *item.Message
		}
		link := ""
		if item.Link != nil {
			link = *item.Link
		}
		fmt.Fprintf(w, "%d\t%d\t%t\t%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.UserID,
			item.Unread,
			item.Category,
			item.CreatedAt,
			readAt,
			message,
			link,
		)
	}
	w.Flush()
	return nil
}

type notificationListItem struct {
	ID        int32   `json:"id"`
	UserID    int32   `json:"user_id"`
	Link      *string `json:"link,omitempty"`
	Message   *string `json:"message,omitempty"`
	CreatedAt string  `json:"created_at"`
	ReadAt    *string `json:"read_at,omitempty"`
	Unread    bool    `json:"unread"`
	Category  string  `json:"category"`
}

func newNotificationListItem(n *db.Notification) notificationListItem {
	item := notificationListItem{
		ID:        n.ID,
		UserID:    n.UsersIdusers,
		Link:      nullStringPtr(n.Link),
		Message:   nullStringPtr(n.Message),
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
		Unread:    !n.ReadAt.Valid,
		Category:  notificationCategory(n.Link),
	}
	if n.ReadAt.Valid {
		read := n.ReadAt.Time.Format(time.RFC3339)
		item.ReadAt = &read
	}
	return item
}

func notificationCategory(link sql.NullString) string {
	if !link.Valid {
		return ""
	}
	return notificationCategoryFromLink(link.String)
}

func notificationCategoryFromLink(link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}
	if u, err := url.Parse(link); err == nil {
		if u.Path != "" {
			return notificationCategoryFromPath(u.Path)
		}
	}
	return notificationCategoryFromPath(link)
}

func notificationCategoryFromPath(path string) string {
	path = strings.TrimLeft(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.SplitN(path, "/", 2)
	return parts[0]
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid || value.String == "" {
		return nil
	}
	v := value.String
	return &v
}

// notificationsMarkCmd implements "notifications mark".
type notificationsMarkCmd struct {
	*notificationsCmd
	fs      *flag.FlagSet
	read    bool
	unread  bool
	jsonOut bool
}

func parseNotificationsMarkCmd(parent *notificationsCmd, args []string) (*notificationsMarkCmd, error) {
	c := &notificationsMarkCmd{notificationsCmd: parent}
	fs, _, err := parseFlags("mark", args, func(fs *flag.FlagSet) {
		fs.BoolVar(&c.read, "read", false, "mark notifications as read (default)")
		fs.BoolVar(&c.unread, "unread", false, "mark notifications as unread")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

// Usage prints command usage information with examples.
func (c *notificationsMarkCmd) Usage() {
	executeUsage(c.fs.Output(), "notifications_mark_usage.txt", c)
}

func (c *notificationsMarkCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*notificationsMarkCmd)(nil)

func (c *notificationsMarkCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing notification IDs")
	}
	if c.read && c.unread {
		return fmt.Errorf("cannot combine --read and --unread")
	}
	status := "read"
	if c.unread {
		status = "unread"
	}
	ids := make([]int32, 0, len(args))
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil || id <= 0 {
			return fmt.Errorf("invalid notification ID %q", arg)
		}
		ids = append(ids, int32(id))
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)

	for _, id := range ids {
		if status == "unread" {
			if err := queries.AdminMarkNotificationUnread(ctx, id); err != nil {
				return fmt.Errorf("mark notification unread: %w", err)
			}
		} else {
			if err := queries.AdminMarkNotificationRead(ctx, id); err != nil {
				return fmt.Errorf("mark notification read: %w", err)
			}
		}
	}

	if c.jsonOut {
		out := map[string]any{
			"status": status,
			"ids":    ids,
			"count":  len(ids),
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	fmt.Fprintf(c.fs.Output(), "Marked %d notifications as %s.\n", len(ids), status)
	return nil
}
