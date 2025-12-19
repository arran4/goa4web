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

	// Build a map of topicID -> thread count using stats query (threads should be obtained by query)
	topicCounts, err := queries.AdminForumTopicThreadCounts(ctx)
	if err != nil {
		return fmt.Errorf("getting topic thread counts: %w", err)
	}
	threadsByTopic := make(map[int32]int64, len(topicCounts))
	for _, tc := range topicCounts {
		// Only consider private forum topics
		if tc.Handler == "private" {
			threadsByTopic[tc.Idforumtopic] = tc.Threads
		}
	}

	var deletedTopics []deletedItem
	var totalGrantsDeleted int

	for _, topic := range topics {
		threads := threadsByTopic[topic.Idforumtopic]
		if threads == 0 {
			if c.verbose {
				log.Printf("Found empty private forum topic: %s (ID: %d)", topic.Title, topic.Idforumtopic)
			}

			grants, err := queries.AdminListGrantsByTopicID(ctx, sql.NullInt32{Int32: topic.Idforumtopic, Valid: true})
			if err != nil {
				log.Printf("error getting grants for topic %d: %v", topic.Idforumtopic, err)
				continue
			}

			participantsMap := make(map[string]struct{})
			for _, grant := range grants {
				if grant.Username.Valid {
					participantsMap[grant.Username.String] = struct{}{}
				}
				if c.verbose {
					log.Printf("  - Deleting grant ID: %d", grant.ID)
				}
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

			if c.verbose {
				log.Printf("  - Deleting topic ID: %d", topic.Idforumtopic)
			}
			if !c.dryRun {
				if err := queries.AdminDeleteForumTopic(ctx, topic.Idforumtopic); err != nil {
					log.Printf("  - error deleting topic %d: %v", topic.Idforumtopic, err)
				}
			}
			deletedTopics = append(deletedTopics, deletedItem{
				ID:           topic.Idforumtopic,
				Title:        topic.Title,
				Participants: participants,
				Type:         "Topic",
			})
		}
	}

	printSummary(deletedTopics, totalGrantsDeleted)
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

	for _, thread := range threads {
		// Use PostCount (comments count) from query to determine emptiness
		if !thread.PostCount.Valid || thread.PostCount.Int32 == 0 {
			if c.verbose {
				log.Printf("Found empty private forum thread: (ID: %d)", thread.Idforumthread)
			}

			grants, err := queries.AdminListGrantsByThreadID(ctx, sql.NullInt32{Int32: thread.Idforumthread, Valid: true})
			if err != nil {
				log.Printf("error getting grants for thread %d: %v", thread.Idforumthread, err)
				continue
			}

			participantsMap := make(map[string]struct{})
			for _, grant := range grants {
				if grant.Username.Valid {
					participantsMap[grant.Username.String] = struct{}{}
				}
				if c.verbose {
					log.Printf("  - Deleting grant ID: %d", grant.ID)
				}
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

			if c.verbose {
				log.Printf("  - Deleting thread ID: %d", thread.Idforumthread)
			}
			if !c.dryRun {
				if err := queries.AdminDeleteForumThread(ctx, thread.Idforumthread); err != nil {
					log.Printf("  - error deleting thread %d: %v", thread.Idforumthread, err)
				}
			}

			var title string
			if t, ok := thread.Title.(string); ok {
				title = t
			} else if tBytes, ok := thread.Title.([]byte); ok {
				title = string(tBytes)
			} else {
				title = fmt.Sprintf("%v", thread.Title)
			}

			deletedThreads = append(deletedThreads, deletedItem{
				ID:           thread.Idforumthread,
				Title:        title,
				Participants: participants,
				Type:         "Thread",
			})
		}
	}

	printSummary(deletedThreads, totalGrantsDeleted)
	log.Println("Cleanup of empty private forum threads complete.")
	return nil
}

func printSummary(items []deletedItem, grantsDeleted int) {
	if len(items) == 0 {
		fmt.Println("No items were affected.")
		return
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("Total items affected: %d\n", len(items))
	fmt.Printf("Total grants deleted: %d\n\n", grantsDeleted)

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
