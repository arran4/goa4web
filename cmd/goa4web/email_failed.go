package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/workers/emailqueue"
)

// emailFailedCmd handles failed email subcommands.
type emailFailedCmd struct {
	*emailCmd
	fs *flag.FlagSet
}

func parseEmailFailedCmd(parent *emailCmd, args []string) (*emailFailedCmd, error) {
	c := &emailFailedCmd{emailCmd: parent}
	c.fs = newFlagSet("failed")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailFailedCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing failed command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseEmailFailedListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "resend":
		cmd, err := parseEmailFailedResendCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("resend: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown failed command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *emailFailedCmd) Usage() {
	executeUsage(c.fs.Output(), "email_failed_usage.txt", c)
}

func (c *emailFailedCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*emailFailedCmd)(nil)

// emailFailedListCmd implements "email failed list".
type emailFailedListCmd struct {
	*emailFailedCmd
	fs       *flag.FlagSet
	Limit    int
	Offset   int
	LangID   int
	Role     string
	Provider string
	Status   string
}

func parseEmailFailedListCmd(parent *emailFailedCmd, args []string) (*emailFailedListCmd, error) {
	c := &emailFailedListCmd{emailFailedCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.IntVar(&c.Limit, "limit", 50, "number of results to return")
	c.fs.IntVar(&c.Offset, "offset", 0, "number of filtered results to skip")
	c.fs.IntVar(&c.LangID, "lang", 0, "language id filter")
	c.fs.StringVar(&c.Role, "role", "", "role name filter")
	c.fs.StringVar(&c.Provider, "provider", "", "provider filter (direct, user, userless)")
	c.fs.StringVar(&c.Status, "status", "failed", "status filter (failed)")

	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

type emailFailedItem struct {
	row      *db.AdminListFailedEmailsRow
	email    string
	subject  string
	provider string
}

func (c *emailFailedListCmd) Run() error {
	if c.Limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	if c.Offset < 0 {
		return fmt.Errorf("offset must be zero or positive")
	}
	status := strings.ToLower(strings.TrimSpace(c.Status))
	if status == "" {
		status = "failed"
	}
	if status != "failed" {
		return fmt.Errorf("unsupported status %q", c.Status)
	}
	provider := strings.ToLower(strings.TrimSpace(c.Provider))
	if provider != "" && provider != "direct" && provider != "user" && provider != "userless" {
		return fmt.Errorf("unsupported provider %q", c.Provider)
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	params := db.AdminListFailedEmailsParams{
		LanguageID: sqlNullInt(c.LangID),
		RoleName:   c.Role,
	}
	pageRows, totalCount, err := c.loadFailedEmails(ctx, queries, params, provider)
	if err != nil {
		return err
	}
	items, err := c.decorateFailedEmails(ctx, queries, pageRows)
	if err != nil {
		return err
	}
	out := c.fs.Output()
	fmt.Fprintf(out, "Failed emails: %d\n", totalCount)
	fmt.Fprintf(out, "Showing %d (offset %d, limit %d)\n\n", len(items), c.Offset, c.Limit)
	fmt.Fprintln(out, "ID\tProvider\tEmail\tSubject\tCreated\tErrors")
	for _, item := range items {
		fmt.Fprintf(out, "%d\t%s\t%s\t%s\t%s\t%d\n",
			item.row.ID,
			item.provider,
			item.email,
			item.subject,
			item.row.CreatedAt.Format(time.RFC3339),
			item.row.ErrorCount,
		)
	}
	return nil
}

func (c *emailFailedListCmd) loadFailedEmails(ctx context.Context, queries *db.Queries, params db.AdminListFailedEmailsParams, provider string) ([]*db.AdminListFailedEmailsRow, int, error) {
	// chunkSize controls how many rows to scan per page while counting totals.
	const chunkSize = 200
	params.Limit = chunkSize
	params.Offset = 0
	var pageRows []*db.AdminListFailedEmailsRow
	filteredIndex := 0
	for {
		rows, err := queries.AdminListFailedEmails(ctx, params)
		if err != nil {
			return nil, 0, fmt.Errorf("list failed emails: %w", err)
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			if provider != "" && provider != emailFailedProvider(row) {
				continue
			}
			if filteredIndex >= c.Offset && len(pageRows) < c.Limit+1 {
				pageRows = append(pageRows, row)
			}
			filteredIndex++
		}
		if len(rows) < chunkSize {
			break
		}
		params.Offset += int32(len(rows))
	}
	if len(pageRows) > c.Limit {
		pageRows = pageRows[:c.Limit]
	}
	return pageRows, filteredIndex, nil
}

func (c *emailFailedListCmd) decorateFailedEmails(ctx context.Context, queries *db.Queries, rows []*db.AdminListFailedEmailsRow) ([]emailFailedItem, error) {
	ids := make([]int32, 0, len(rows))
	for _, row := range rows {
		if row.ToUserID.Valid {
			ids = append(ids, row.ToUserID.Int32)
		}
	}
	users := make(map[int32]*db.SystemGetUserByIDRow)
	for _, id := range ids {
		if u, err := queries.SystemGetUserByID(ctx, id); err == nil {
			users[id] = u
		}
	}
	items := make([]emailFailedItem, 0, len(rows))
	for _, row := range rows {
		emailStr := ""
		if row.ToUserID.Valid && !row.DirectEmail {
			if u, ok := users[row.ToUserID.Int32]; ok && u.Email.Valid && u.Email.String != "" {
				emailStr = u.Email.String
			}
		}
		subject := ""
		if m, err := mail.ReadMessage(strings.NewReader(row.Body)); err == nil {
			if emailStr == "" {
				emailStr = m.Header.Get("To")
			}
			subject = m.Header.Get("Subject")
		}
		if emailStr == "" {
			emailStr = "(unknown)"
		}
		if row.DirectEmail {
			emailStr += " (direct)"
		} else if !row.ToUserID.Valid || row.ToUserID.Int32 == 0 {
			emailStr += " (userless)"
		}
		items = append(items, emailFailedItem{
			row:      row,
			email:    emailStr,
			subject:  subject,
			provider: emailFailedProvider(row),
		})
	}
	return items, nil
}

func emailFailedProvider(row *db.AdminListFailedEmailsRow) string {
	if row.DirectEmail {
		return "direct"
	}
	if row.ToUserID.Valid && row.ToUserID.Int32 != 0 {
		return "user"
	}
	return "userless"
}

type intList []int

func (l *intList) String() string {
	if l == nil {
		return ""
	}
	vals := make([]string, 0, len(*l))
	for _, v := range *l {
		vals = append(vals, strconv.Itoa(v))
	}
	return strings.Join(vals, ",")
}

func (l *intList) Set(value string) error {
	id, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid id %q", value)
	}
	*l = append(*l, id)
	return nil
}

// emailFailedResendCmd implements "email failed resend".
type emailFailedResendCmd struct {
	*emailFailedCmd
	fs  *flag.FlagSet
	ids intList
}

func parseEmailFailedResendCmd(parent *emailFailedCmd, args []string) (*emailFailedResendCmd, error) {
	c := &emailFailedResendCmd{emailFailedCmd: parent}
	c.fs = newFlagSet("resend")
	c.fs.Var(&c.ids, "id", "failed email id (repeatable)")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	for _, arg := range c.fs.Args() {
		if err := c.ids.Set(arg); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *emailFailedResendCmd) Run() error {
	if len(c.ids) == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	provider, err := c.rootCmd.emailReg.ProviderFromConfig(c.rootCmd.cfg)
	if err != nil {
		return fmt.Errorf("email provider: %w", err)
	}
	var emails []*db.AdminGetPendingEmailByIDRow
	for _, id := range c.ids {
		email, err := queries.AdminGetPendingEmailByID(ctx, int32(id))
		if err != nil {
			return fmt.Errorf("get email: %w", err)
		}
		emails = append(emails, email)
	}
	for _, email := range emails {
		addr, err := emailqueue.ResolveQueuedEmailAddress(ctx, queries, c.rootCmd.cfg, &db.SystemListPendingEmailsRow{ID: email.ID, ToUserID: email.ToUserID, Body: email.Body, ErrorCount: email.ErrorCount, DirectEmail: email.DirectEmail})
		if err != nil {
			return fmt.Errorf("resolve address: %w", err)
		}
		if provider != nil {
			if err := provider.Send(ctx, addr, []byte(email.Body)); err != nil {
				return fmt.Errorf("send email: %w", err)
			}
		}
		if err := queries.SystemMarkPendingEmailSent(ctx, email.ID); err != nil {
			return fmt.Errorf("mark sent: %w", err)
		}
	}
	return nil
}

func sqlNullInt(value int) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(value), Valid: value != 0}
}
