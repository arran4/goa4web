package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// newsReadCmd implements "news read".
type newsReadCmd struct {
	*newsCmd
	fs   *flag.FlagSet
	ID   int
	args []string
}

func parseNewsReadCmd(parent *newsCmd, args []string) (*newsReadCmd, error) {
	c := &newsReadCmd{newsCmd: parent}
	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "news id")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	if c.ID == 0 && len(c.args) > 0 {
		if id, err := strconv.Atoi(c.args[0]); err == nil {
			c.ID = id
			c.args = c.args[1:]
		}
	}
	return c, nil
}

func (c *newsReadCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	row, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("get news: %w", err)
	}
	if row.Occured.Valid {
		fmt.Printf("Posted: %s\n", row.Occured.Time.Format(time.RFC3339))
	}
	fmt.Println(row.News.String)
	return nil
}
