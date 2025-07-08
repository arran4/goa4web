package main

import (
	"context"
	"flag"
	"fmt"
	"net/mail"
	"strings"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// emailQueueListCmd implements "email queue list".
type emailQueueListCmd struct {
	*emailQueueCmd
	fs   *flag.FlagSet
	args []string
}

func parseEmailQueueListCmd(parent *emailQueueCmd, args []string) (*emailQueueListCmd, error) {
	c := &emailQueueListCmd{emailQueueCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *emailQueueListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.ListUnsentPendingEmails(ctx)
	if err != nil {
		return fmt.Errorf("list emails: %w", err)
	}
	ids := make([]int32, 0, len(rows))
	for _, e := range rows {
		ids = append(ids, e.ToUserID)
	}
	users, err := queries.UsersByID(ctx, ids)
	if err != nil {
		return fmt.Errorf("get users: %w", err)
	}
	for _, e := range rows {
		emailStr := ""
		if u, ok := users[e.ToUserID]; ok && u.Email.Valid {
			emailStr = u.Email.String
		}
		subj := ""
		if m, err := mail.ReadMessage(strings.NewReader(e.Body)); err == nil {
			subj = m.Header.Get("Subject")
		}
		fmt.Printf("%d\t%s\t%s\t%s\n", e.ID, emailStr, subj, e.CreatedAt.Format(time.RFC3339))
	}
	return nil
}
