package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// faqReadCmd implements "faq read".
type faqReadCmd struct {
	*faqCmd
	fs   *flag.FlagSet
	ID   int
	args []string
}

func parseFaqReadCmd(parent *faqCmd, args []string) (*faqReadCmd, error) {
	c := &faqReadCmd{faqCmd: parent}
	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "faq id")
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

func (c *faqReadCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.GetAllFAQQuestions(ctx)
	if err != nil {
		return fmt.Errorf("get faq: %w", err)
	}
	for _, q := range rows {
		if int(q.Idfaq) == c.ID {
			fmt.Printf("Q: %s\n", q.Question.String)
			fmt.Printf("A: %s\n", q.Answer.String)
			return nil
		}
	}
	return fmt.Errorf("faq not found")
}
