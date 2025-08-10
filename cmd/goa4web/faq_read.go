package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/internal/db"
)

// faqReadCmd implements "faq read".
type faqReadCmd struct {
	*faqCmd
	fs *flag.FlagSet
	ID int
}

func parseFaqReadCmd(parent *faqCmd, args []string) (*faqReadCmd, error) {
	c := &faqReadCmd{faqCmd: parent}
	c.fs = newFlagSet("read")
	c.fs.IntVar(&c.ID, "id", 0, "faq id")
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

func (c *faqReadCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.SystemGetFAQQuestions(ctx)
	if err != nil {
		return fmt.Errorf("get faq: %w", err)
	}
	for _, q := range rows {
		if int(q.ID) == c.ID {
			fmt.Printf("Q: %s\n", q.Question.String)
			fmt.Printf("A: %s\n", q.Answer.String)
			return nil
		}
	}
	return fmt.Errorf("faq not found")
}
