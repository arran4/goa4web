package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// blogCreateCmd implements "blog create".
type blogCreateCmd struct {
	*blogCmd
	fs     *flag.FlagSet
	UserID int
	LangID int
	Text   string
}

func parseBlogCreateCmd(parent *blogCmd, args []string) (*blogCreateCmd, error) {
	c := &blogCreateCmd{blogCmd: parent}
	c.fs = newFlagSet("create")
	c.fs.IntVar(&c.UserID, "user", 0, "user id")
	c.fs.IntVar(&c.LangID, "lang", 0, "language id")
	c.fs.StringVar(&c.Text, "text", "", "blog text")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *blogCreateCmd) Run() error {
	if c.UserID == 0 || c.LangID == 0 || c.Text == "" {
		return fmt.Errorf("user, lang and text required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	_, err = queries.CreateBlogEntryForWriter(ctx, db.CreateBlogEntryForWriterParams{
		UsersIdusers:       int32(c.UserID),
		LanguageIdlanguage: int32(c.LangID),
		Blog:               sql.NullString{String: c.Text, Valid: true},
		UserID:             sql.NullInt32{Int32: int32(c.UserID), Valid: true},
		ListerID:           int32(c.UserID),
	})
	if err != nil {
		return fmt.Errorf("create blog: %w", err)
	}
	return nil
}
