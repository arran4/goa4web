package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

const (
	// commentListDeactivatedFormatTable is the default output format for deactivated comment listings.
	commentListDeactivatedFormatTable = "table"
	// commentListDeactivatedFormatJSON is the JSON output format for deactivated comment listings.
	commentListDeactivatedFormatJSON = "json"
	// commentListDeactivatedDateLayout is the YYYY-MM-DD layout for date filters.
	commentListDeactivatedDateLayout = "2006-01-02"
	// commentListDeactivatedTypeForum is the filter value for forum comments.
	commentListDeactivatedTypeForum = "forum"
	// commentListDeactivatedTypeNews is the filter value for news comments.
	commentListDeactivatedTypeNews = "news"
	// commentListDeactivatedTypeBlog is the filter value for blog comments.
	commentListDeactivatedTypeBlog = "blog"
)

// commentListDeactivatedCmd implements "comment list-deactivated".
type commentListDeactivatedCmd struct {
	*commentCmd
	fs          *flag.FlagSet
	Limit       int
	Offset      int
	ContentType string
	DateFrom    string
	DateTo      string
	Format      string
}

func parseCommentListDeactivatedCmd(parent *commentCmd, args []string) (*commentListDeactivatedCmd, error) {
	c := &commentListDeactivatedCmd{commentCmd: parent}
	fs, _, err := parseFlags("list-deactivated", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.Limit, "limit", 20, "max results")
		fs.IntVar(&c.Offset, "offset", 0, "result offset")
		fs.StringVar(&c.ContentType, "type", "", "filter by content type (news, blog, forum)")
		fs.StringVar(&c.DateFrom, "from", "", "filter deactivated after (YYYY-MM-DD or RFC3339)")
		fs.StringVar(&c.DateTo, "to", "", "filter deactivated before (YYYY-MM-DD or RFC3339)")
		fs.StringVar(&c.Format, "format", commentListDeactivatedFormatTable, "output format (table, json)")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *commentListDeactivatedCmd) Run() error {
	format := strings.ToLower(strings.TrimSpace(c.Format))
	if format == "" {
		format = commentListDeactivatedFormatTable
	}
	if format != commentListDeactivatedFormatTable && format != commentListDeactivatedFormatJSON {
		return fmt.Errorf("invalid format %q (expected %s or %s)", c.Format, commentListDeactivatedFormatTable, commentListDeactivatedFormatJSON)
	}
	filterType, err := normalizeCommentContentTypeFilter(c.ContentType)
	if err != nil {
		return err
	}
	from, fromSet, err := parseCommentListDateFilter(c.DateFrom, false)
	if err != nil {
		return err
	}
	to, toSet, err := parseCommentListDateFilter(c.DateTo, true)
	if err != nil {
		return err
	}
	if fromSet && toSet && from.After(to) {
		return fmt.Errorf("from date is after to date")
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.AdminListDeactivatedComments(ctx, db.AdminListDeactivatedCommentsParams{Limit: int32(c.Limit), Offset: int32(c.Offset)})
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}

	type commentListItem struct {
		ID          int32  `json:"id"`
		ThreadID    int32  `json:"thread_id"`
		ContentType string `json:"content_type"`
		WrittenAt   string `json:"written_at,omitempty"`
		DeletedAt   string `json:"deleted_at,omitempty"`
		Text        string `json:"text"`
	}

	items := make([]commentListItem, 0, len(rows))
	for _, r := range rows {
		details, err := queries.AdminGetDeactivatedCommentById(ctx, r.Idcomments)
		if err != nil {
			return fmt.Errorf("load comment %d: %w", r.Idcomments, err)
		}
		contentType, err := commentContentTypeFromThread(ctx, queries, details.ForumthreadID)
		if err != nil {
			return fmt.Errorf("thread %d: %w", details.ForumthreadID, err)
		}
		if filterType != "" && contentType != filterType {
			continue
		}
		if (fromSet || toSet) && !commentDeletedWithin(details.DeletedAt, from, fromSet, to, toSet) {
			continue
		}

		item := commentListItem{
			ID:          details.Idcomments,
			ThreadID:    details.ForumthreadID,
			ContentType: contentType,
			Text:        details.Text.String,
		}
		if details.Written.Valid {
			item.WrittenAt = details.Written.Time.Format(time.RFC3339)
		}
		if details.DeletedAt.Valid {
			item.DeletedAt = details.DeletedAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	if format == commentListDeactivatedFormatJSON {
		payload, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return fmt.Errorf("json output: %w", err)
		}
		fmt.Println(string(payload))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tThreadID\tType\tWritten\tDeactivated\tText")
	for _, item := range items {
		fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.ThreadID,
			item.ContentType,
			item.WrittenAt,
			item.DeletedAt,
			item.Text,
		)
	}
	w.Flush()
	return nil
}

func normalizeCommentContentTypeFilter(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case commentListDeactivatedTypeForum:
		return commentListDeactivatedTypeForum, nil
	case commentListDeactivatedTypeNews:
		return commentListDeactivatedTypeNews, nil
	case commentListDeactivatedTypeBlog, "blogs":
		return commentListDeactivatedTypeBlog, nil
	default:
		return "", fmt.Errorf("invalid content type %q (expected %s, %s, or %s)", value, commentListDeactivatedTypeNews, commentListDeactivatedTypeBlog, commentListDeactivatedTypeForum)
	}
}

func parseCommentListDateFilter(value string, isEnd bool) (time.Time, bool, error) {
	if value == "" {
		return time.Time{}, false, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, true, nil
	}
	t, err := time.Parse(commentListDeactivatedDateLayout, value)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("invalid date %q (expected RFC3339 or %s)", value, commentListDeactivatedDateLayout)
	}
	if isEnd {
		return t.Add(24*time.Hour - time.Nanosecond), true, nil
	}
	return t, true, nil
}

func commentDeletedWithin(deletedAt sql.NullTime, from time.Time, fromSet bool, to time.Time, toSet bool) bool {
	if !deletedAt.Valid {
		return !fromSet && !toSet
	}
	ts := deletedAt.Time
	if fromSet && ts.Before(from) {
		return false
	}
	if toSet && ts.After(to) {
		return false
	}
	return true
}

func commentContentTypeFromThread(ctx context.Context, queries *db.Queries, threadID int32) (string, error) {
	if threadID == 0 {
		return "unknown", nil
	}
	thread, err := queries.AdminGetForumThreadById(ctx, threadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "unknown", nil
		}
		return "", err
	}
	switch thread.TopicHandler {
	case "":
		return commentListDeactivatedTypeForum, nil
	case "news":
		return commentListDeactivatedTypeNews, nil
	case "blogs":
		return commentListDeactivatedTypeBlog, nil
	default:
		return thread.TopicHandler, nil
	}
}
