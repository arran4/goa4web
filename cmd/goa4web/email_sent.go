package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/mail"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// emailSentCmd handles email sent subcommands.
type emailSentCmd struct {
	*emailCmd
	fs *flag.FlagSet
}

func parseEmailSentCmd(parent *emailCmd, args []string) (*emailSentCmd, error) {
	c := &emailSentCmd{emailCmd: parent}
	c.fs = newFlagSet("sent")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailSentCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing sent command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseEmailSentListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "retry":
		cmd, err := parseEmailSentRetryCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("retry: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown sent command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *emailSentCmd) Usage() {
	executeUsage(c.fs.Output(), "email_sent_usage.txt", c)
}

func (c *emailSentCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*emailSentCmd)(nil)

// emailSentListCmd implements "email sent list".
type emailSentListCmd struct {
	*emailSentCmd
	fs        *flag.FlagSet
	recipient string
	provider  string
	from      string
	to        string
	limit     int
	offset    int
}

type sentEmailItem struct {
	ID          int32  `json:"id"`
	ToUserID    *int32 `json:"to_user_id,omitempty"`
	Recipient   string `json:"recipient"`
	Subject     string `json:"subject"`
	Provider    string `json:"provider"`
	CreatedAt   string `json:"created_at"`
	SentAt      string `json:"sent_at"`
	ErrorCount  int32  `json:"error_count"`
	DirectEmail bool   `json:"direct_email"`
}

type sentEmailListOutput struct {
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	Total  int             `json:"total"`
	Items  []sentEmailItem `json:"items"`
}

func parseEmailSentListCmd(parent *emailSentCmd, args []string) (*emailSentListCmd, error) {
	c := &emailSentListCmd{emailSentCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.Usage = c.Usage
	c.fs.StringVar(&c.recipient, "recipient", "", "filter by recipient email address or user ID")
	c.fs.StringVar(&c.provider, "provider", "", "filter by provider: direct, user, userless, or any")
	c.fs.StringVar(&c.from, "from", "", "filter by sent time starting at this RFC3339 timestamp")
	c.fs.StringVar(&c.to, "to", "", "filter by sent time before or at this RFC3339 timestamp")
	c.fs.IntVar(&c.limit, "limit", 50, "number of results to return")
	c.fs.IntVar(&c.offset, "offset", 0, "starting offset for results")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailSentListCmd) Run() error {
	if c.limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}
	if c.offset < 0 {
		return fmt.Errorf("offset must be zero or greater")
	}
	filterProvider, err := parseProviderFilter(c.provider)
	if err != nil {
		return err
	}
	fromTime, err := parseTimestamp(c.from)
	if err != nil {
		return err
	}
	toTime, err := parseTimestamp(c.to)
	if err != nil {
		return err
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)

	rows, err := queries.AdminListSentEmails(ctx, db.AdminListSentEmailsParams{
		LanguageID: sql.NullInt32{},
		RoleName:   "",
		Limit:      int32(maxQueryLimit),
		Offset:     0,
	})
	if err != nil {
		return fmt.Errorf("list sent emails: %w", err)
	}

	filtered, err := c.filterSentEmails(ctx, queries, rows, filterProvider, fromTime, toTime)
	if err != nil {
		return err
	}

	start := c.offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + c.limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return c.renderListOutput(filtered[start:end], len(filtered))
}

func parseProviderFilter(raw string) (string, error) {
	if raw == "" || strings.EqualFold(raw, "any") {
		return "", nil
	}
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case "direct", "user", "userless":
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid provider %q", raw)
	}
}

func (c *emailSentListCmd) filterSentEmails(ctx context.Context, queries *db.Queries, rows []*db.AdminListSentEmailsRow, provider string, fromTime *time.Time, toTime *time.Time) ([]sentEmailItem, error) {
	items := make([]sentEmailItem, 0, len(rows))
	userIDs := make([]int32, 0, len(rows))
	for _, row := range rows {
		if row.ToUserID.Valid && row.ToUserID.Int32 > 0 {
			userIDs = append(userIDs, row.ToUserID.Int32)
		}
	}
	userEmails, err := lookupUserEmails(ctx, queries, userIDs)
	if err != nil {
		return nil, err
	}
	recipientFilter := strings.TrimSpace(c.recipient)
	for _, row := range rows {
		if !row.SentAt.Valid {
			continue
		}
		if fromTime != nil && row.SentAt.Time.Before(*fromTime) {
			continue
		}
		if toTime != nil && row.SentAt.Time.After(*toTime) {
			continue
		}
		rowProvider := resolveSentProvider(row.DirectEmail, row.ToUserID)
		if provider != "" && rowProvider != provider {
			continue
		}
		recipient, subject := resolveRecipientAndSubject(row.Body, row.ToUserID, row.DirectEmail, userEmails)
		if recipientFilter != "" && !matchesRecipientFilter(recipientFilter, recipient, row.ToUserID) {
			continue
		}
		item := sentEmailItem{
			ID:          row.ID,
			Recipient:   recipient,
			Subject:     subject,
			Provider:    rowProvider,
			CreatedAt:   row.CreatedAt.Format(time.RFC3339),
			SentAt:      row.SentAt.Time.Format(time.RFC3339),
			ErrorCount:  row.ErrorCount,
			DirectEmail: row.DirectEmail,
		}
		if row.ToUserID.Valid {
			id := row.ToUserID.Int32
			item.ToUserID = &id
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].SentAt == items[j].SentAt {
			return items[i].ID > items[j].ID
		}
		return items[i].SentAt > items[j].SentAt
	})
	return items, nil
}

func resolveSentProvider(direct bool, toUserID sql.NullInt32) string {
	if direct {
		return "direct"
	}
	if toUserID.Valid && toUserID.Int32 != 0 {
		return "user"
	}
	return "userless"
}

func resolveRecipientAndSubject(body string, toUserID sql.NullInt32, direct bool, userEmails map[int32]string) (string, string) {
	recipient := ""
	if toUserID.Valid && !direct {
		if email, ok := userEmails[toUserID.Int32]; ok && email != "" {
			recipient = email
		}
	}
	subject := ""
	if msg, err := mail.ReadMessage(strings.NewReader(body)); err == nil {
		if recipient == "" {
			recipient = msg.Header.Get("To")
		}
		subject = msg.Header.Get("Subject")
	}
	if recipient == "" {
		recipient = "(unknown)"
	}
	return recipient, subject
}

func matchesRecipientFilter(filter string, recipient string, toUserID sql.NullInt32) bool {
	if strings.EqualFold(filter, recipient) {
		return true
	}
	if toUserID.Valid {
		if id, err := strconv.Atoi(filter); err == nil && int32(id) == toUserID.Int32 {
			return true
		}
	}
	return false
}

func lookupUserEmails(ctx context.Context, queries *db.Queries, ids []int32) (map[int32]string, error) {
	if len(ids) == 0 {
		return map[int32]string{}, nil
	}
	unique := make(map[int32]struct{}, len(ids))
	for _, id := range ids {
		if id > 0 {
			unique[id] = struct{}{}
		}
	}
	keys := make([]int32, 0, len(unique))
	for id := range unique {
		keys = append(keys, id)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	results := make(map[int32]string, len(keys))
	for _, id := range keys {
		user, err := queries.SystemGetUserByID(ctx, id)
		if err != nil {
			continue
		}
		if user.Email.Valid && user.Email.String != "" {
			results[id] = user.Email.String
		}
	}
	return results, nil
}

func (c *emailSentListCmd) renderListOutput(items []sentEmailItem, total int) error {
	output := sentEmailListOutput{
		Limit:  c.limit,
		Offset: c.offset,
		Total:  total,
		Items:  items,
	}
	b, _ := json.MarshalIndent(output, "", "  ")
	fmt.Fprintln(c.fs.Output(), string(b))
	return nil
}

// emailSentRetryCmd implements "email sent retry".
type emailSentRetryCmd struct {
	*emailSentCmd
	fs  *flag.FlagSet
	id  int
	ids string
}

type sentRetryOutput struct {
	RetriedIDs []int32 `json:"retried_ids"`
	Count      int     `json:"count"`
}

func parseEmailSentRetryCmd(parent *emailSentCmd, args []string) (*emailSentRetryCmd, error) {
	c := &emailSentRetryCmd{emailSentCmd: parent}
	c.fs = newFlagSet("retry")
	c.fs.Usage = c.Usage
	c.fs.IntVar(&c.id, "id", 0, "sent email id")
	c.fs.StringVar(&c.ids, "ids", "", "comma-separated list of sent email ids")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailSentRetryCmd) Run() error {
	ids, err := c.parseIDs()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	for _, id := range ids {
		email, err := queries.AdminGetPendingEmailByID(ctx, id)
		if err != nil {
			return fmt.Errorf("get email: %w", err)
		}
		if err := queries.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToUserID: email.ToUserID, Body: email.Body, DirectEmail: email.DirectEmail}); err != nil {
			return fmt.Errorf("queue email: %w", err)
		}
	}
	out := sentRetryOutput{RetriedIDs: ids, Count: len(ids)}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Fprintln(c.fs.Output(), string(b))
	return nil
}

func (c *emailSentRetryCmd) parseIDs() ([]int32, error) {
	ids := make([]int32, 0)
	if c.id != 0 {
		ids = append(ids, int32(c.id))
	}
	if c.ids == "" {
		return ids, nil
	}
	parts := strings.Split(c.ids, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("parse id %q: %w", part, err)
		}
		ids = append(ids, int32(val))
	}
	return ids, nil
}
