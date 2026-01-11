package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

type forumThreadCmd struct {
	*forumCmd
	fs *flag.FlagSet
}

func parseForumThreadCmd(parent *forumCmd, args []string) (*forumThreadCmd, error) {
	c := &forumThreadCmd{forumCmd: parent}
	c.fs = newFlagSet("thread")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *forumThreadCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing thread command")
	}
	switch args[0] {
	case "list":
		return c.runList(args[1:])
	case "details":
		return c.runDetails(args[1:])
	case "delete":
		return c.runDelete(args[1:])
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown thread command %q", args[0])
	}
}

func (c *forumThreadCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage: %s forum thread <command> [flags]\n", os.Args[0])
	fmt.Fprintln(c.fs.Output(), "\nCommands:")
	fmt.Fprintln(c.fs.Output(), "  list    List forum threads")
	fmt.Fprintln(c.fs.Output(), "  details Show details of a forum thread")
	fmt.Fprintln(c.fs.Output(), "  delete  Delete a forum thread")
}

func (c *forumThreadCmd) runList(args []string) error {
	fs := newFlagSet("list")
	limit := fs.Int("limit", 20, "Number of threads to list")
	offset := fs.Int("offset", 0, "Offset for listing threads")
	if err := fs.Parse(args); err != nil {
		return err
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return err
	}
	queries := db.New(conn)
	ctx := context.Background()

	threads, err := queries.AdminListForumThreads(ctx, db.AdminListForumThreadsParams{
		Limit:  int32(*limit),
		Offset: int32(*offset),
	})
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTopicID\tTopic\tTitle\tPosts\tLast Post")
	for _, t := range threads {
		if isPublicForumHandler(t.TopicHandler) {
			fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%d\t%v\n", t.Idforumthread, t.Idforumtopic, t.TopicTitle.String, t.Title, t.PostCount.Int32, t.LastPostAt.Time)
		}
	}
	w.Flush()
	return nil
}

func (c *forumThreadCmd) runDetails(args []string) error {
	fs := newFlagSet("details")
	id := fs.Int("id", 0, "Thread ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == 0 {
		return fmt.Errorf("missing -id")
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return err
	}
	queries := db.New(conn)
	ctx := context.Background()

	thread, err := queries.AdminGetForumThreadById(ctx, int32(*id))
	if err != nil {
		return err
	}

	fmt.Printf("ID: %d\n", thread.Idforumthread)
	fmt.Printf("Topic ID: %d\n", thread.Idforumtopic)
	fmt.Printf("Topic: %s\n", thread.TopicTitle.String)
	fmt.Printf("Handler: %s\n", thread.TopicHandler)
	fmt.Printf("Title: %s\n", thread.Title)
	fmt.Printf("Created At: %v\n", thread.CreatedAt.Time)
	fmt.Printf("Created By User ID: %d\n", thread.CreatedBy)
	fmt.Printf("Last Post By User ID: %d\n", thread.LastPostBy)
	fmt.Printf("Last Post At: %v\n", thread.LastPostAt.Time)
	fmt.Printf("Post Count: %d\n", thread.PostCount.Int32)
	return nil
}

func (c *forumThreadCmd) runDelete(args []string) error {
	fs := newFlagSet("delete")
	id := fs.Int("id", 0, "Thread ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == 0 {
		return fmt.Errorf("missing -id")
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return err
	}
	queries := db.New(conn)
	ctx := context.Background()

	// Verification it is public thread
	thread, err := queries.AdminGetForumThreadById(ctx, int32(*id))
	if err != nil {
		return err
	}
	if !isPublicForumHandler(thread.TopicHandler) {
		return fmt.Errorf("thread %d is not a public forum thread", *id)
	}

	return queries.AdminDeleteForumThread(ctx, int32(*id))
}
