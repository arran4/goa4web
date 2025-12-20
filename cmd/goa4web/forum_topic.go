package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

type forumTopicCmd struct {
	*forumCmd
	fs *flag.FlagSet
}

func parseForumTopicCmd(parent *forumCmd, args []string) (*forumTopicCmd, error) {
	c := &forumTopicCmd{forumCmd: parent}
	c.fs = newFlagSet("topic")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *forumTopicCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing topic command")
	}
	switch args[0] {
	case "list":
		return c.runList(args[1:])
	case "details":
		return c.runDetails(args[1:])
	case "delete":
		return c.runDelete(args[1:])
	case "edit":
		return c.runEdit(args[1:])
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown topic command %q", args[0])
	}
}

func (c *forumTopicCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage: %s forum topic <command> [flags]\n", os.Args[0])
	fmt.Fprintln(c.fs.Output(), "\nCommands:")
	fmt.Fprintln(c.fs.Output(), "  list    List forum topics")
	fmt.Fprintln(c.fs.Output(), "  details Show details of a forum topic")
	fmt.Fprintln(c.fs.Output(), "  delete  Delete a forum topic")
	fmt.Fprintln(c.fs.Output(), "  edit    Edit a forum topic")
}

func (c *forumTopicCmd) runList(args []string) error {
	fs := newFlagSet("list")
	limit := fs.Int("limit", 20, "Number of topics to list")
	offset := fs.Int("offset", 0, "Offset for listing topics")
	if err := fs.Parse(args); err != nil {
		return err
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return err
	}
	queries := db.New(conn)
	ctx := context.Background()

	topics, err := queries.AdminListForumTopics(ctx, db.AdminListForumTopicsParams{
		Limit:  int32(*limit),
		Offset: int32(*offset),
	})
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tHandler\tThreads\tComments")
	for _, t := range topics {
		if isPublicForumHandler(t.Handler) {
			fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\n", t.Idforumtopic, t.Title.String, t.Handler, t.Threads.Int32, t.Comments.Int32)
		}
	}
	w.Flush()
	return nil
}

func (c *forumTopicCmd) runDetails(args []string) error {
	fs := newFlagSet("details")
	id := fs.Int("id", 0, "Topic ID")
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

	topic, err := queries.GetForumTopicById(ctx, int32(*id))
	if err != nil {
		return err
	}

	fmt.Printf("ID: %d\n", topic.Idforumtopic)
	fmt.Printf("Title: %s\n", topic.Title.String)
	fmt.Printf("Description: %s\n", topic.Description.String)
	fmt.Printf("Handler: %s\n", topic.Handler)
	fmt.Printf("Category ID: %d\n", topic.ForumcategoryIdforumcategory)
	fmt.Printf("Language ID: %d\n", topic.LanguageID.Int32)
	fmt.Printf("Threads: %d\n", topic.Threads.Int32)
	fmt.Printf("Comments: %d\n", topic.Comments.Int32)
	fmt.Printf("Last Addition: %v\n", topic.Lastaddition.Time)
	return nil
}

func (c *forumTopicCmd) runDelete(args []string) error {
	fs := newFlagSet("delete")
	id := fs.Int("id", 0, "Topic ID")
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

	// Check if topic exists and is public
	topic, err := queries.GetForumTopicById(ctx, int32(*id))
	if err != nil {
		return err
	}
	if !isPublicForumHandler(topic.Handler) {
		return fmt.Errorf("topic %d is not a public forum topic (handler: %s)", *id, topic.Handler)
	}

	return queries.AdminDeleteForumTopic(ctx, int32(*id))
}

func (c *forumTopicCmd) runEdit(args []string) error {
	fs := newFlagSet("edit")
	id := fs.Int("id", 0, "Topic ID")
	title := fs.String("title", "", "New title")
	description := fs.String("description", "", "New description")
	categoryID := fs.Int("category-id", 0, "New category ID")
	languageID := fs.Int("language-id", 0, "New language ID (0 for null)")

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

	topic, err := queries.GetForumTopicById(ctx, int32(*id))
	if err != nil {
		return err
	}
	if !isPublicForumHandler(topic.Handler) {
		return fmt.Errorf("topic %d is not a public forum topic", *id)
	}

	newTitle := topic.Title.String
	if *title != "" {
		newTitle = *title
	}
	newDesc := topic.Description.String
	if *description != "" {
		newDesc = *description
	}
	newCatID := topic.ForumcategoryIdforumcategory
	if *categoryID != 0 {
		newCatID = int32(*categoryID)
	}
	newLangID := topic.LanguageID
	if isFlagPassed(fs, "language-id") {
		if *languageID == 0 {
			newLangID = sql.NullInt32{Valid: false}
		} else {
			newLangID = sql.NullInt32{Int32: int32(*languageID), Valid: true}
		}
	}

	return queries.AdminUpdateForumTopic(ctx, db.AdminUpdateForumTopicParams{
		Title:                        sql.NullString{String: newTitle, Valid: true},
		Description:                  sql.NullString{String: newDesc, Valid: true},
		ForumcategoryIdforumcategory: newCatID,
		TopicLanguageID:              newLangID,
		Idforumtopic:                 topic.Idforumtopic,
	})
}
