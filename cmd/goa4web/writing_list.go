package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/core/common"
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
}

func parseWritingListCmd(parent *writingCmd, args []string) (*writingListCmd, error) {
	c := &writingListCmd{writingCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.IntVar(&c.UserID, "user", 0, "user id")
	c.fs.IntVar(&c.Category, "category", 0, "category id")
	c.fs.IntVar(&c.Limit, "limit", 10, "limit")
	c.fs.IntVar(&c.Offset, "offset", 0, "offset")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

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
			WritingCategoryID: int32(c.Category),
			Limit:             int32(c.Limit),
			Offset:            int32(c.Offset),
		})
		if err != nil {
			return fmt.Errorf("list writings: %w", err)
		}
		for _, w := range rows {
			fmt.Printf("%d\t%s\n", w.Idwriting, w.Title.String)
		}
		return nil
	}
	cd := common.NewCoreData(ctx, queries)
	rows, err := cd.LatestWritings(common.WithWritingsOffset(int32(c.Offset)), common.WithWritingsLimit(int32(c.Limit)))
	if err != nil {
		return fmt.Errorf("list writings: %w", err)
	}
	for _, w := range rows {
		fmt.Printf("%d\t%s\n", w.Idwriting, w.Title.String)
	}
	return nil
}
