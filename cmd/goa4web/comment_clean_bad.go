package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

// commentCleanBadCmd implements "comment clean-bad".
type commentCleanBadCmd struct {
	*commentCmd
	fs      *flag.FlagSet
	dryRun  bool
	verbose bool
}

func parseCommentCleanBadCmd(parent *commentCmd, args []string) (*commentCleanBadCmd, error) {
	c := &commentCleanBadCmd{commentCmd: parent}
	c.fs = newFlagSet("clean-bad")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "Dry run")
	c.fs.BoolVar(&c.verbose, "verbose", false, "Verbose output")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *commentCleanBadCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	queries := db.New(conn)
	ctx := context.Background()

	if c.dryRun {
		log.Println("Dry run mode enabled. No changes will be made.")
	}

	comments, err := queries.AdminListBadComments(ctx)
	if err != nil {
		return fmt.Errorf("getting bad comments: %w", err)
	}

	var deletedComments []*db.Comment

	for _, comment := range comments {
		if c.verbose {
			log.Printf("Found bad comment ID: %d", comment.Idcomments)
		}

		if !c.dryRun {
			if err := queries.AdminHardDeleteComment(ctx, comment.Idcomments); err != nil {
				log.Printf("  - error deleting comment %d: %v", comment.Idcomments, err)
				continue
			}
		}

		deletedComments = append(deletedComments, comment)
	}

	c.printSummary(deletedComments)
	log.Println("Cleanup of bad comments complete.")
	return nil
}

func (c *commentCleanBadCmd) printSummary(items []*db.Comment) {
	if len(items) == 0 {
		fmt.Println("No items were affected.")
		return
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("Total items affected: %d\n\n", len(items))

	w := tabwriter.NewWriter(c.fs.Output(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tThreadID\tUserID\tWritten")
	fmt.Fprintln(w, "--\t--------\t------\t-------")

	for _, item := range items {
		written := "NULL"
		if item.Written.Valid {
			written = item.Written.Time.Format("2006-01-02 15:04:05")
		}
		fmt.Fprintf(w, "%d\t%v\t%d\t%s\n", item.Idcomments, item.ForumthreadID, item.UsersIdusers, written)
	}
	w.Flush()
	fmt.Println()
}
