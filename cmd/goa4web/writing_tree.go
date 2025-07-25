package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
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
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.FetchAllCategories(ctx)
	if err != nil {
		return fmt.Errorf("tree: %w", err)
	}
	children := map[int32][]*dbpkg.WritingCategory{}
	for _, cat := range rows {
		parent := cat.WritingCategoryID
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
