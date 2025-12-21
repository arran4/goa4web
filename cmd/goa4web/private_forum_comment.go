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

type privateForumCommentCmd struct {
	*privateForumCmd
	fs *flag.FlagSet
}

func parsePrivateForumCommentCmd(parent *privateForumCmd, args []string) (*privateForumCommentCmd, error) {
	c := &privateForumCommentCmd{privateForumCmd: parent}
	c.fs = newFlagSet("comment")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *privateForumCommentCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing comment command")
	}
	switch args[0] {
	case "list":
		return c.runList(args[1:])
	case "details":
		return c.runDetails(args[1:])
	case "delete":
		return c.runDelete(args[1:])
	case "deactivate":
		return c.runDeactivate(args[1:])
	case "activate":
		return c.runActivate(args[1:])
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comment command %q", args[0])
	}
}

func (c *privateForumCommentCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage: %s private-forum comment <command> [flags]\n", os.Args[0])
	fmt.Fprintln(c.fs.Output(), "\nCommands:")
	fmt.Fprintln(c.fs.Output(), "  list        List private forum comments")
	fmt.Fprintln(c.fs.Output(), "  details     Show details of a comment")
	fmt.Fprintln(c.fs.Output(), "  delete      Permanently delete a comment")
	fmt.Fprintln(c.fs.Output(), "  deactivate  Deactivate (soft delete) a comment")
	fmt.Fprintln(c.fs.Output(), "  activate    Activate (restore) a comment")
}

func (c *privateForumCommentCmd) runList(args []string) error {
	fs := newFlagSet("list")
	limit := fs.Int("limit", 20, "Number of comments to list")
	offset := fs.Int("offset", 0, "Offset for listing comments")
	if err := fs.Parse(args); err != nil {
		return err
	}

	conn, err := c.rootCmd.DB()
	if err != nil {
		return err
	}
	queries := db.New(conn)
	ctx := context.Background()

	comments, err := queries.AdminListPrivateForumComments(ctx, db.AdminListPrivateForumCommentsParams{
		Limit:  int32(*limit),
		Offset: int32(*offset),
	})
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tThreadID\tUser\tDate\tText\tDeleted")
	for _, com := range comments {
		text := com.Text.String
		if len(text) > 50 {
			text = text[:47] + "..."
		}
		deleted := ""
		if com.DeletedAt.Valid {
			deleted = "YES"
		}
		fmt.Fprintf(w, "%d\t%d\t%s\t%v\t%s\t%s\n", com.Idcomments, com.Idforumthread.Int32, com.Posterusername.String, com.Written.Time, text, deleted)
	}
	w.Flush()
	return nil
}

func (c *privateForumCommentCmd) runDetails(args []string) error {
	fs := newFlagSet("details")
	id := fs.Int("id", 0, "Comment ID")
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

	comment, err := queries.GetCommentById(ctx, int32(*id))
	if err != nil {
		return err
	}

	// Verify scope
	thread, err := queries.AdminGetForumThreadById(ctx, comment.ForumthreadID)
	if err != nil {
		return fmt.Errorf("failed to get thread info for comment: %w", err)
	}
	if !isPrivateForumHandler(thread.TopicHandler) {
		return fmt.Errorf("comment %d is not in a private forum thread", *id)
	}

	fmt.Printf("ID: %d\n", comment.Idcomments)
	fmt.Printf("Thread ID: %d\n", comment.ForumthreadID)
	fmt.Printf("User ID: %d\n", comment.UsersIdusers)
	fmt.Printf("Written: %v\n", comment.Written.Time)
	fmt.Printf("Deleted At: %v\n", comment.DeletedAt.Time)
	fmt.Printf("Text:\n%s\n", comment.Text.String)
	return nil
}

func (c *privateForumCommentCmd) runDelete(args []string) error {
	fs := newFlagSet("delete")
	id := fs.Int("id", 0, "Comment ID")
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

	comment, err := queries.GetCommentById(ctx, int32(*id))
	if err != nil {
		return err
	}

	thread, err := queries.AdminGetForumThreadById(ctx, comment.ForumthreadID)
	if err != nil {
		return fmt.Errorf("failed to get thread info for comment: %w", err)
	}

	if !isPrivateForumHandler(thread.TopicHandler) {
		return fmt.Errorf("comment %d is not in a private forum thread", *id)
	}

	return queries.AdminHardDeleteComment(ctx, int32(*id))
}

func (c *privateForumCommentCmd) runDeactivate(args []string) error {
	fs := newFlagSet("deactivate")
	id := fs.Int("id", 0, "Comment ID")
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

	// Check scope
	comment, err := queries.GetCommentById(ctx, int32(*id))
	if err != nil {
		return err
	}
	thread, err := queries.AdminGetForumThreadById(ctx, comment.ForumthreadID)
	if err != nil {
		return err
	}
	if !isPrivateForumHandler(thread.TopicHandler) {
		return fmt.Errorf("comment %d is not in a private forum thread", *id)
	}

	// Archive first
	if err := queries.AdminArchiveComment(ctx, db.AdminArchiveCommentParams{
		Idcomments:    comment.Idcomments,
		ForumthreadID: comment.ForumthreadID,
		UsersIdusers:  comment.UsersIdusers,
		LanguageID:    comment.LanguageID,
		Written:       comment.Written,
		Text:          comment.Text,
		Timezone:      comment.Timezone,
	}); err != nil {
		return fmt.Errorf("failed to archive comment: %w", err)
	}

	// Scrub (soft delete)
	return queries.AdminScrubComment(ctx, db.AdminScrubCommentParams{
		Text:       sql.NullString{String: "[Deleted]", Valid: true},
		Idcomments: int32(*id),
	})
}

func (c *privateForumCommentCmd) runActivate(args []string) error {
	fs := newFlagSet("activate")
	id := fs.Int("id", 0, "Comment ID")
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

	// Get archived content
	archived, err := queries.AdminGetDeactivatedCommentById(ctx, int32(*id))
	if err != nil {
		return fmt.Errorf("failed to find deactivated comment %d: %w", *id, err)
	}

	// Check scope (thread handler)
	thread, err := queries.AdminGetForumThreadById(ctx, archived.ForumthreadID)
	if err != nil {
		return fmt.Errorf("failed to get thread info: %w", err)
	}
	if !isPrivateForumHandler(thread.TopicHandler) {
		return fmt.Errorf("comment %d is not in a private forum thread", *id)
	}

	// Restore
	if err := queries.AdminRestoreComment(ctx, db.AdminRestoreCommentParams{
		Text:       archived.Text,
		Idcomments: archived.Idcomments,
	}); err != nil {
		return fmt.Errorf("failed to restore comment: %w", err)
	}

	// Mark restored in deactivated table
	return queries.AdminMarkCommentRestored(ctx, int32(*id))
}
