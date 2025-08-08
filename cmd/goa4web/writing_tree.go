package main

import (
	"context"
	"flag"
	"fmt"
	"math"

	"github.com/arran4/goa4web/internal/db"
)

// writingTreeCmd implements "writing tree".
type writingTreeCmd struct {
	*writingCmd
	fs *flag.FlagSet
}

func parseWritingTreeCmd(parent *writingCmd, args []string) (*writingTreeCmd, error) {
	c := &writingTreeCmd{writingCmd: parent}
	c.fs = newFlagSet("tree")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *writingTreeCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.SystemListWritingCategories(ctx, db.SystemListWritingCategoriesParams{Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		return fmt.Errorf("tree: %w", err)
	}
	children := map[int32][]*db.WritingCategory{}
	for _, cat := range rows {
		parent := int32(0)
		if cat.WritingCategoryID.Valid {
			parent = cat.WritingCategoryID.Int32
		}
		children[parent] = append(children[parent], cat)
	}
	var printTree func(parent int32, prefix string)
	printTree = func(parent int32, prefix string) {
		for _, cat := range children[parent] {
			fmt.Printf("%s%d\t%s\n", prefix, cat.Idwritingcategory, cat.Title.String)
			printTree(cat.Idwritingcategory, prefix+"  ")
		}
	}
	printTree(0, "")
	return nil
}
