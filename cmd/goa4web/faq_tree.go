package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// faqTreeCmd implements "faq tree".
type faqTreeCmd struct {
	*faqCmd
	fs   *flag.FlagSet
	args []string
}

func parseFaqTreeCmd(parent *faqCmd, args []string) (*faqTreeCmd, error) {
	c := &faqTreeCmd{faqCmd: parent}
	fs := flag.NewFlagSet("tree", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *faqTreeCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.GetAllAnsweredFAQWithFAQCategories(ctx)
	if err != nil {
		return fmt.Errorf("tree: %w", err)
	}
	var lastCat int32 = -1
	for _, r := range rows {
		if r.Idfaqcategories.Valid && r.Idfaqcategories.Int32 != lastCat {
			fmt.Printf("%d\t%s\n", r.Idfaqcategories.Int32, r.Name.String)
			lastCat = r.Idfaqcategories.Int32
		}
		fmt.Printf("  %d\t%s\n", r.Idfaq, r.Question.String)
	}
	return nil
}
