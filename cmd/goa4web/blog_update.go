package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// blogUpdateCmd implements "blog update".
type blogUpdateCmd struct {
	*blogCmd
	fs     *flag.FlagSet
	ID     int
	LangID int
	Text   string
}

func parseBlogUpdateCmd(parent *blogCmd, args []string) (*blogUpdateCmd, error) {
	c := &blogUpdateCmd{blogCmd: parent}
	c.fs = newFlagSet("update")
	c.fs.IntVar(&c.ID, "id", 0, "blog id")
	c.fs.IntVar(&c.LangID, "lang", 0, "language id")
	c.fs.StringVar(&c.Text, "text", "", "blog text")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *blogUpdateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	err = queries.UpdateBlogEntry(ctx, dbpkg.UpdateBlogEntryParams{
		LanguageIdlanguage: int32(c.LangID),
		Blog:               sql.NullString{String: c.Text, Valid: c.Text != ""},
		Idblogs:            int32(c.ID),
	})
	if err != nil {
		return fmt.Errorf("update blog: %w", err)
	}
	return nil
}
