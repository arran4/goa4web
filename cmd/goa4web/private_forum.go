package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

func isPrivateForumHandler(handler string) bool {
	return handler == "private"
}

// privateForumCmd handles private-forum utilities.
type privateForumCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parsePrivateForumCmd(parent *rootCmd, args []string) (*privateForumCmd, error) {
	c := &privateForumCmd{rootCmd: parent}
	c.fs = newFlagSet("private-forum")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *privateForumCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing private-forum command")
	}
	switch args[0] {
	case "clean-empty":
		cmd, err := parsePrivateForumCleanEmptyCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("clean-empty: %w", err)
		}
		return cmd.Run()
	case "clean-empty-threads":
		cmd, err := parsePrivateForumCleanEmptyThreadsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("clean-empty-threads: %w", err)
		}
		return cmd.Run()
	case "topic":
		cmd, err := parsePrivateForumTopicCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("topic: %w", err)
		}
		return cmd.Run()
	case "thread":
		cmd, err := parsePrivateForumThreadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("thread: %w", err)
		}
		return cmd.Run()
	case "comment":
		cmd, err := parsePrivateForumCommentCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("comment: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown private-forum command %q", args[0])
	}
}

// Usage prints command usage information.
func (c *privateForumCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage: %s private-forum <command> [flags]\n\n", c.rootCmd.fs.Name())
	fmt.Fprintln(c.fs.Output(), "Commands:")
	fmt.Fprintln(c.fs.Output(), "  clean-empty          Clean up empty private forum topics")
	fmt.Fprintln(c.fs.Output(), "  clean-empty-threads  Clean up empty private forum threads")
	fmt.Fprintln(c.fs.Output(), "  topic                Manage private forum topics")
	fmt.Fprintln(c.fs.Output(), "  thread               Manage private forum threads")
	fmt.Fprintln(c.fs.Output(), "  comment              Manage private forum comments")
}

// privateForumCleanEmptyCmd implements "private-forum clean-empty".
type privateForumCleanEmptyCmd struct {
	*privateForumCmd
	fs      *flag.FlagSet
	dryRun  bool
	verbose bool
}

func parsePrivateForumCleanEmptyCmd(parent *privateForumCmd, args []string) (*privateForumCleanEmptyCmd, error) {
	c := &privateForumCleanEmptyCmd{privateForumCmd: parent}
	c.fs = newFlagSet("clean-empty")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "Dry run")
	c.fs.BoolVar(&c.verbose, "verbose", false, "Verbose output")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

type deletedItem struct {
	ID           int32
	Title        string
	Participants []string
	Type         string
}

func (c *privateForumCleanEmptyCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	queries := db.New(conn)
	ctx := context.Background()

	if c.dryRun {
		log.Println("Dry run mode enabled. No changes will be made.")
	}

	topics, err := queries.AdminListAllPrivateTopics(ctx)
	if err != nil {
		return fmt.Errorf("getting private forum topics: %w", err)
	}

	threads, err := queries.AdminListAllPrivateForumThreads(ctx)
	if err != nil {
		return fmt.Errorf("getting private forum threads: %w", err)
	}

	// Build a map of topicID -> thread count using stats query (threads should be obtained by query)
	topicCounts, err := queries.AdminForumTopicThreadCounts(ctx)
	if err != nil {
		return fmt.Errorf("getting topic thread counts: %w", err)
	}
	threadsByTopic := make(map[int32]int64, len(topicCounts))
	for _, tc := range topicCounts {
		// Only consider private forum topics
		if isPrivateForumHandler(tc.Handler) {
			threadsByTopic[tc.Idforumtopic] = tc.Threads
		}
	}

	threadStatsByTopic := make(map[int32][]*db.AdminListAllPrivateForumThreadsRow)
	for _, thread := range threads {
		threadStatsByTopic[thread.Idforumtopic] = append(threadStatsByTopic[thread.Idforumtopic], thread)
	}

	var deletedItems []deletedItem
	var totalGrantsDeleted int
	var totalCommentsDeleted int

	for _, topic := range topics {
		threadCount := threadsByTopic[topic.Idforumtopic]
		threadsForTopic := threadStatsByTopic[topic.Idforumtopic]
		validThreads := 0
		for _, thread := range threadsForTopic {
			if thread.ValidComments > 0 {
				validThreads++
			}
		}

		deleteReason := ""
		if threadCount == 0 {
			deleteReason = "no threads"
		}

		if deleteReason == "" && validThreads == 0 {
			deleteReason = "no content"
		}

		if deleteReason == "" {
			continue
		}

		participantsMap := make(map[string]struct{})

		log.Printf("Deleting private forum topic ID %d (%s): %s", topic.Idforumtopic, topic.Title, deleteReason)

		for _, thread := range threadsForTopic {
			threadTitle := formatThreadTitle(thread.Title)
			threadParticipants, grantsDeleted, commentsDeleted := c.deletePrivateForumThread(ctx, queries, thread, threadTitle, "topic deletion")
			totalGrantsDeleted += grantsDeleted
			totalCommentsDeleted += commentsDeleted
			deletedItems = append(deletedItems, deletedItem{
				ID:           thread.Idforumthread,
				Title:        threadTitle,
				Participants: threadParticipants,
				Type:         "Thread",
			})
		}

		grants, err := queries.AdminListGrantsByTopicID(ctx, sql.NullInt32{Int32: topic.Idforumtopic, Valid: true})
		if err != nil {
			log.Printf("error getting grants for topic %d: %v", topic.Idforumtopic, err)
			continue
		}

		for _, grant := range grants {
			if grant.Username.Valid {
				participantsMap[grant.Username.String] = struct{}{}
			}
			log.Printf("  - Deleting topic grant ID: %d", grant.ID)
			if !c.dryRun {
				if err := queries.AdminDeleteGrant(ctx, grant.ID); err != nil {
					log.Printf("  - error deleting grant %d: %v", grant.ID, err)
				}
			}
			totalGrantsDeleted++
		}

		var participants []string
		for p := range participantsMap {
			participants = append(participants, p)
		}
		sort.Strings(participants)

		log.Printf("  - Deleting topic ID: %d", topic.Idforumtopic)
		if !c.dryRun {
			if err := queries.AdminDeleteForumTopic(ctx, topic.Idforumtopic); err != nil {
				log.Printf("  - error deleting topic %d: %v", topic.Idforumtopic, err)
			}
		}
		deletedItems = append(deletedItems, deletedItem{
			ID:           topic.Idforumtopic,
			Title:        topic.Title,
			Participants: participants,
			Type:         "Topic",
		})
	}

	printSummary(deletedItems, totalGrantsDeleted, totalCommentsDeleted)
	log.Println("Cleanup of empty private forum topics complete.")
	return nil
}

// privateForumCleanEmptyThreadsCmd implements "private-forum clean-empty-threads".
type privateForumCleanEmptyThreadsCmd struct {
	*privateForumCmd
	fs      *flag.FlagSet
	dryRun  bool
	verbose bool
}

func parsePrivateForumCleanEmptyThreadsCmd(parent *privateForumCmd, args []string) (*privateForumCleanEmptyThreadsCmd, error) {
	c := &privateForumCleanEmptyThreadsCmd{privateForumCmd: parent}
	c.fs = newFlagSet("clean-empty-threads")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "Dry run")
	c.fs.BoolVar(&c.verbose, "verbose", false, "Verbose output")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *privateForumCleanEmptyThreadsCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	queries := db.New(conn)
	ctx := context.Background()

	if c.dryRun {
		log.Println("Dry run mode enabled. No changes will be made.")
	}

	threads, err := queries.AdminListAllPrivateForumThreads(ctx)
	if err != nil {
		return fmt.Errorf("getting private forum threads: %w", err)
	}

	var deletedThreads []deletedItem
	var totalGrantsDeleted int
	var totalCommentsDeleted int

	for _, thread := range threads {
		threadTitle := formatThreadTitle(thread.Title)
		if thread.ValidComments == 0 {
			reason := "no comments"
			if thread.InvalidComments > 0 {
				reason = "only invalid comments"
			}
			log.Printf("Deleting private forum thread ID %d (%s): %s", thread.Idforumthread, threadTitle, reason)

			participants, grantsDeleted, commentsDeleted := c.deletePrivateForumThread(ctx, queries, thread, threadTitle, reason)
			totalGrantsDeleted += grantsDeleted
			totalCommentsDeleted += commentsDeleted
			deletedThreads = append(deletedThreads, deletedItem{
				ID:           thread.Idforumthread,
				Title:        threadTitle,
				Participants: participants,
				Type:         "Thread",
			})
			continue
		}

		if thread.InvalidComments > 0 {
			log.Printf("Cleaning invalid comments for thread ID %d (%s)", thread.Idforumthread, threadTitle)
			commentsDeleted, err := c.deletePrivateForumThreadInvalidComments(ctx, queries, thread.Idforumthread)
			if err != nil {
				log.Printf("  - error deleting invalid comments for thread %d: %v", thread.Idforumthread, err)
				continue
			}
			totalCommentsDeleted += commentsDeleted
		}
	}

	printSummary(deletedThreads, totalGrantsDeleted, totalCommentsDeleted)
	log.Println("Cleanup of empty private forum threads complete.")
	return nil
}

func (c *privateForumCleanEmptyThreadsCmd) deletePrivateForumThreadInvalidComments(ctx context.Context, queries *db.Queries, threadID int32) (int, error) {
	comments, err := queries.AdminListPrivateForumInvalidCommentsByThread(ctx, threadID)
	if err != nil {
		return 0, err
	}

	if len(comments) == 0 {
		return 0, nil
	}

	for _, commentID := range comments {
		_, err := queries.AdminGetSubsequentCommentID(ctx, db.AdminGetSubsequentCommentIDParams{
			ForumthreadID: threadID,
			Idcomments:    commentID,
		})
		if err == nil {
			log.Printf("  - Reporting invalid comment ID %d: has subsequent comments", commentID)
			continue
		}

		log.Printf("  - Deleting invalid comment ID: %d", commentID)
		if !c.dryRun {
			if err := queries.AdminHardDeleteComment(ctx, commentID); err != nil {
				log.Printf("  - error deleting comment %d: %v", commentID, err)
				continue
			}
		}
	}

	return len(comments), nil
}

func (c *privateForumCleanEmptyCmd) deletePrivateForumThread(ctx context.Context, queries *db.Queries, thread *db.AdminListAllPrivateForumThreadsRow, title string, reason string) ([]string, int, int) {
	log.Printf("  - Deleting thread ID: %d (%s): %s", thread.Idforumthread, title, reason)

	if thread.TotalComments > 0 {
		log.Printf("  - Deleting %d comments for thread ID: %d", thread.TotalComments, thread.Idforumthread)
	}
	if !c.dryRun {
		if err := queries.AdminDeleteCommentsByThread(ctx, thread.Idforumthread); err != nil {
			log.Printf("  - error deleting comments for thread %d: %v", thread.Idforumthread, err)
		}
	}

	grants, err := queries.AdminListGrantsByThreadID(ctx, sql.NullInt32{Int32: thread.Idforumthread, Valid: true})
	if err != nil {
		log.Printf("error getting grants for thread %d: %v", thread.Idforumthread, err)
		return nil, 0, int(thread.TotalComments)
	}

	participantsMap := make(map[string]struct{})
	for _, grant := range grants {
		if grant.Username.Valid {
			participantsMap[grant.Username.String] = struct{}{}
		}
		log.Printf("  - Deleting thread grant ID: %d", grant.ID)
		if !c.dryRun {
			if err := queries.AdminDeleteGrant(ctx, grant.ID); err != nil {
				log.Printf("  - error deleting grant %d: %v", grant.ID, err)
			}
		}
	}

	log.Printf("  - Deleting thread ID: %d", thread.Idforumthread)
	if !c.dryRun {
		if err := queries.AdminDeleteForumThread(ctx, thread.Idforumthread); err != nil {
			log.Printf("  - error deleting thread %d: %v", thread.Idforumthread, err)
		}
	}

	var participants []string
	for p := range participantsMap {
		participants = append(participants, p)
	}
	sort.Strings(participants)

	return participants, len(grants), int(thread.TotalComments)
}

func (c *privateForumCleanEmptyThreadsCmd) deletePrivateForumThread(ctx context.Context, queries *db.Queries, thread *db.AdminListAllPrivateForumThreadsRow, title string, reason string) ([]string, int, int) {
	log.Printf("  - Deleting thread ID: %d (%s): %s", thread.Idforumthread, title, reason)

	if thread.TotalComments > 0 {
		log.Printf("  - Deleting %d comments for thread ID: %d", thread.TotalComments, thread.Idforumthread)
	}
	if !c.dryRun {
		if err := queries.AdminDeleteCommentsByThread(ctx, thread.Idforumthread); err != nil {
			log.Printf("  - error deleting comments for thread %d: %v", thread.Idforumthread, err)
		}
	}

	grants, err := queries.AdminListGrantsByThreadID(ctx, sql.NullInt32{Int32: thread.Idforumthread, Valid: true})
	if err != nil {
		log.Printf("error getting grants for thread %d: %v", thread.Idforumthread, err)
		return nil, 0, int(thread.TotalComments)
	}

	participantsMap := make(map[string]struct{})
	for _, grant := range grants {
		if grant.Username.Valid {
			participantsMap[grant.Username.String] = struct{}{}
		}
		log.Printf("  - Deleting thread grant ID: %d", grant.ID)
		if !c.dryRun {
			if err := queries.AdminDeleteGrant(ctx, grant.ID); err != nil {
				log.Printf("  - error deleting grant %d: %v", grant.ID, err)
			}
		}
	}

	log.Printf("  - Deleting thread ID: %d", thread.Idforumthread)
	if !c.dryRun {
		if err := queries.AdminDeleteForumThread(ctx, thread.Idforumthread); err != nil {
			log.Printf("  - error deleting thread %d: %v", thread.Idforumthread, err)
		}
	}

	var participants []string
	for p := range participantsMap {
		participants = append(participants, p)
	}
	sort.Strings(participants)

	return participants, len(grants), int(thread.TotalComments)
}

func formatThreadTitle(title interface{}) string {
	switch value := title.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return fmt.Sprintf("%v", title)
	}
}

func printSummary(items []deletedItem, grantsDeleted int, commentsDeleted int) {
	if len(items) == 0 && grantsDeleted == 0 && commentsDeleted == 0 {
		fmt.Println("No items were affected.")
		return
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("Total items affected: %d\n", len(items))
	fmt.Printf("Total grants deleted: %d\n\n", grantsDeleted)
	fmt.Printf("Total comments deleted: %d\n\n", commentsDeleted)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tType\tTitle\tParticipants")
	fmt.Fprintln(w, "--\t----\t-----\t------------")

	for _, item := range items {
		participants := strings.Join(item.Participants, ", ")
		if participants == "" {
			participants = "(none)"
		}
		title := item.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}
		// Escape tabs/newlines in title just in case
		title = strings.ReplaceAll(title, "\t", " ")
		title = strings.ReplaceAll(title, "\n", " ")
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", item.ID, item.Type, title, participants)
	}
	w.Flush()
	fmt.Println()
}
