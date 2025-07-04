package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// blogCreateCmd implements "blog create".
type blogCreateCmd struct {
	*blogCmd
	fs     *flag.FlagSet
	UserID int
	LangID int
	Text   string
	args   []string
}

func parseBlogCreateCmd(parent *blogCmd, args []string) (*blogCreateCmd, error) {
	c := &blogCreateCmd{blogCmd: parent}
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.IntVar(&c.UserID, "user", 0, "user id")
	fs.IntVar(&c.LangID, "lang", 0, "language id")
	fs.StringVar(&c.Text, "text", "", "blog text")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *blogCreateCmd) Run() error {
	if c.UserID == 0 || c.LangID == 0 || c.Text == "" {
		return fmt.Errorf("user, lang and text required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	_, err = queries.CreateBlogEntry(ctx, dbpkg.CreateBlogEntryParams{
		UsersIdusers:       int32(c.UserID),
		LanguageIdlanguage: int32(c.LangID),
		Blog:               sql.NullString{String: c.Text, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create blog: %w", err)
	}
	return nil
}
