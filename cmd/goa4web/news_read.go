package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// newsReadCmd implements "news read".
type newsReadCmd struct {
	*newsCmd
	fs *flag.FlagSet
	ID int
}

func parseNewsReadCmd(parent *newsCmd, args []string) (*newsReadCmd, error) {
	c := &newsReadCmd{newsCmd: parent}
	c.fs = newFlagSet("read")
	c.fs.IntVar(&c.ID, "id", 0, "news id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	rest := c.fs.Args()
	if c.ID == 0 && len(rest) > 0 {
		if id, err := strconv.Atoi(rest[0]); err == nil {
			c.ID = id
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
	queries := db.New(db)
	row, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx, db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: 0,
		ID:       int32(c.ID),
		UserID:   sql.NullInt32{},
	})
	if err != nil {
		return fmt.Errorf("get news: %w", err)
	}
	if row.Occurred.Valid {
		fmt.Printf("Posted: %s\n", row.Occurred.Time.Format(time.RFC3339))
	}
	fmt.Println(row.News.String)
	return nil
}
