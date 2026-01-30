package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

const maxQueryLimit = int64(^uint32(0) >> 1) // max int32 value for query limits.

// userPasswordListCmd implements "user password list".
type userPasswordListCmd struct {
	*userPasswordCmd
	fs     *flag.FlagSet
	page   int
	limit  int
	status string
	user   string
	since  string
	json   bool
}

type passwordResetListItem struct {
	ID         int32   `json:"id"`
	UserID     int32   `json:"user_id"`
	Username   string  `json:"username"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
	VerifiedAt *string `json:"verified_at,omitempty"`
}

type passwordResetListOutput struct {
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
	Total int                     `json:"total"`
	Items []passwordResetListItem `json:"items"`
}

func parseUserPasswordListCmd(parent *userPasswordCmd, args []string) (*userPasswordListCmd, error) {
	c := &userPasswordListCmd{userPasswordCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.Usage = c.Usage
	c.fs.IntVar(&c.page, "page", 1, "page number (starting at 1)")
	c.fs.IntVar(&c.limit, "limit", 20, "number of results per page")
	c.fs.StringVar(&c.status, "status", "pending", "status filter: pending, verified, or all")
	c.fs.StringVar(&c.user, "user", "", "username or user ID to filter")
	c.fs.StringVar(&c.since, "since", "", "only include resets created at or after this RFC3339 timestamp")
	c.fs.BoolVar(&c.json, "json", false, "machine-readable JSON output")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userPasswordListCmd) Run() error {
	if c.page < 1 {
		return fmt.Errorf("page must be at least 1")
	}
	if c.limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}

	statusFilter, err := c.parseStatus()
	if err != nil {
		return err
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)

	userID, err := c.resolveUserID(ctx, queries)
	if err != nil {
		return err
	}

	sinceTime, err := parseTimestamp(c.since)
	if err != nil {
		return err
	}

	count, err := queries.AdminCountPasswordResets(ctx, db.AdminCountPasswordResetsParams{
		Status:        statusFilter,
		UserID:        sql.NullInt32{},
		CreatedBefore: sql.NullTime{},
	})
	if err != nil {
		return fmt.Errorf("count password resets: %w", err)
	}
	if count == 0 {
		return c.renderOutput(nil, nil)
	}

	limit := int64(count)
	if limit > maxQueryLimit {
		limit = maxQueryLimit
	}

	rows, err := queries.AdminListPasswordResets(ctx, db.AdminListPasswordResetsParams{
		Status:        statusFilter,
		UserID:        sql.NullInt32{},
		CreatedBefore: sql.NullTime{},
		Limit:         int32(limit),
		Offset:        0,
	})
	if err != nil {
		return fmt.Errorf("list password resets: %w", err)
	}

	filtered := make([]*db.AdminListPasswordResetsRow, 0, len(rows))
	for _, row := range rows {
		if userID != nil && row.UserID != *userID {
			continue
		}
		if sinceTime != nil && row.CreatedAt.Before(*sinceTime) {
			continue
		}
		filtered = append(filtered, row)
	}

	start := (c.page - 1) * c.limit
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + c.limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return c.renderOutput(filtered[start:end], filtered)
}

func (c *userPasswordListCmd) parseStatus() (sql.NullString, error) {
	status := strings.ToLower(strings.TrimSpace(c.status))
	if status == "" || status == "all" {
		return sql.NullString{}, nil
	}
	if status != "pending" && status != "verified" {
		return sql.NullString{}, fmt.Errorf("invalid status %q", c.status)
	}
	return sql.NullString{String: status, Valid: true}, nil
}

func (c *userPasswordListCmd) resolveUserID(ctx context.Context, queries *db.Queries) (*int32, error) {
	if c.user == "" {
		return nil, nil
	}
	if id, err := strconv.Atoi(c.user); err == nil {
		if id <= 0 {
			return nil, fmt.Errorf("invalid user id %q", c.user)
		}
		userID := int32(id)
		if _, err := queries.SystemGetUserByID(ctx, userID); err != nil {
			return nil, fmt.Errorf("get user by id: %w", err)
		}
		return &userID, nil
	}

	user, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.user, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	userID := user.Idusers
	return &userID, nil
}

func parseTimestamp(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp %q", raw)
		}
	}
	return &parsed, nil
}

func (c *userPasswordListCmd) renderOutput(pageRows []*db.AdminListPasswordResetsRow, allRows []*db.AdminListPasswordResetsRow) error {
	if c.json {
		output := passwordResetListOutput{Page: c.page, Limit: c.limit}
		if allRows != nil {
			output.Total = len(allRows)
		}
		output.Items = make([]passwordResetListItem, 0, len(pageRows))
		for _, row := range pageRows {
			item := passwordResetListItem{
				ID:        row.ID,
				UserID:    row.UserID,
				Username:  toNullString(row.Username),
				Status:    resetStatus(row.VerifiedAt),
				CreatedAt: row.CreatedAt.Format(time.RFC3339),
			}
			if row.VerifiedAt.Valid {
				verifiedAt := row.VerifiedAt.Time.Format(time.RFC3339)
				item.VerifiedAt = &verifiedAt
			}
			output.Items = append(output.Items, item)
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tUser ID\tUsername\tStatus\tCreated At\tVerified At")
	for _, row := range pageRows {
		verifiedAt := "-"
		if row.VerifiedAt.Valid {
			verifiedAt = row.VerifiedAt.Time.Format(time.RFC3339)
		}
		fmt.Fprintf(
			w,
			"%d\t%d\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.UserID,
			toNullString(row.Username),
			resetStatus(row.VerifiedAt),
			row.CreatedAt.Format(time.RFC3339),
			verifiedAt,
		)
	}
	return w.Flush()
}

func resetStatus(verifiedAt sql.NullTime) string {
	if verifiedAt.Valid {
		return "verified"
	}
	return "pending"
}

func toNullString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func (c *userPasswordListCmd) Usage() {
	executeUsage(c.fs.Output(), "user_password_list_usage.txt", c)
}

func (c *userPasswordListCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userPasswordListCmd)(nil)
