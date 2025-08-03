package main

import (
	"context"
	"flag"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// emailQueueListCmd implements "email queue list".
type emailQueueListCmd struct {
	*emailQueueCmd
	fs *flag.FlagSet
}

func parseEmailQueueListCmd(parent *emailQueueCmd, args []string) (*emailQueueListCmd, error) {
	c := &emailQueueListCmd{emailQueueCmd: parent}
	c.fs = newFlagSet("list")

	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailQueueListCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.AdminListUnsentPendingEmails(ctx, db.AdminListUnsentPendingEmailsParams{})
	if err != nil {
		return fmt.Errorf("list emails: %w", err)
	}
	ids := make([]int32, 0, len(rows))
	for _, e := range rows {
		if e.ToUserID.Valid {
			ids = append(ids, e.ToUserID.Int32)
		}
	}
	users := make(map[int32]*db.SystemGetUserByIDRow)
	for _, id := range ids {
		if u, err := queries.SystemGetUserByID(ctx, id); err == nil {
			users[id] = u
		}
	}
	for _, e := range rows {
		emailStr := ""
		if e.ToUserID.Valid {
			if u, ok := users[e.ToUserID.Int32]; ok && u.Email.Valid && u.Email.String != "" {
				emailStr = u.Email.String
			}
		}
		subj := ""
		if m, err := mail.ReadMessage(strings.NewReader(e.Body)); err == nil {
			subj = m.Header.Get("Subject")
		}
		fmt.Printf("%d\t%s\t%s\t%s\n", e.ID, emailStr, subj, e.CreatedAt.Format(time.RFC3339))
	}
	return nil
}
