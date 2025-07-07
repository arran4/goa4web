package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// writingListCmd implements "writing list".
type writingListCmd struct {
	*writingCmd
	fs       *flag.FlagSet
	UserID   int
	Category int
	Limit    int
	Offset   int
	args     []string
}

func parseWritingListCmd(parent *writingCmd, args []string) (*writingListCmd, error) {
	c := &writingListCmd{writingCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.IntVar(&c.UserID, "user", 0, "user id")
	fs.IntVar(&c.Category, "category", 0, "category id")
	fs.IntVar(&c.Limit, "limit", 10, "limit")
	fs.IntVar(&c.Offset, "offset", 0, "offset")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *writingListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if c.UserID != 0 {
		rows, err := queries.GetPublicWritingsByUser(ctx, dbpkg.GetPublicWritingsByUserParams{
			UsersIdusers: int32(c.UserID),
			Limit:        int32(c.Limit),
			Offset:       int32(c.Offset),
		})
		if err != nil {
			return fmt.Errorf("list writings: %w", err)
		}
		for _, w := range rows {
			fmt.Printf("%d\t%s\n", w.Idwriting, w.Title.String)
		}
		return nil
	}
	if c.Category != 0 {
		rows, err := queries.GetPublicWritingsInCategory(ctx, dbpkg.GetPublicWritingsInCategoryParams{
			WritingcategoryIdwritingcategory: int32(c.Category),
			Limit:                            int32(c.Limit),
			Offset:                           int32(c.Offset),
		})
		if err != nil {
			return fmt.Errorf("list writings: %w", err)
		}
		for _, w := range rows {
			fmt.Printf("%d\t%s\n", w.Idwriting, w.Title.String)
		}
		return nil
	}
	rows, err := queries.GetPublicWritings(ctx, dbpkg.GetPublicWritingsParams{
		Limit:  int32(c.Limit),
		Offset: int32(c.Offset),
	})
	if err != nil {
		return fmt.Errorf("list writings: %w", err)
	}
	for _, w := range rows {
		fmt.Printf("%d\t%s\n", w.Idwriting, w.Title.String)
	}
	return nil
}
