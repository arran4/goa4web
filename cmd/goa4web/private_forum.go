package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

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

	for _, topic := range topics {
		threads := threadsByTopic[topic.Idforumtopic]
		if threads == 0 {
			if c.verbose {
				log.Printf("Found empty private forum topic: %s (ID: %d)", topic.Title.String, topic.Idforumtopic)
			}

			grants, err := queries.AdminListGrantsByTopicID(ctx, sql.NullInt32{Int32: topic.Idforumtopic, Valid: true})
			if err != nil {
				log.Printf("error getting grants for topic %d: %v", topic.Idforumtopic, err)
				continue
			}

			for _, grant := range grants {
				if c.verbose {
					log.Printf("  - Deleting grant ID: %d", grant.ID)
				}
				if !c.dryRun {
					if err := queries.AdminDeleteGrant(ctx, grant.ID); err != nil {
						log.Printf("  - error deleting grant %d: %v", grant.ID, err)
					}
				}
			}

			if c.verbose {
				log.Printf("  - Deleting topic ID: %d", topic.Idforumtopic)
			}
			if !c.dryRun {
				if err := queries.AdminDeleteForumTopic(ctx, topic.Idforumtopic); err != nil {
					log.Printf("  - error deleting topic %d: %v", topic.Idforumtopic, err)
				}
			}
		}
	}

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

			for _, grant := range grants {
				if c.verbose {
					log.Printf("  - Deleting grant ID: %d", grant.ID)
				}
				if !c.dryRun {
					if err := queries.AdminDeleteGrant(ctx, grant.ID); err != nil {
						log.Printf("  - error deleting grant %d: %v", grant.ID, err)
					}
				}
			}

			if c.verbose {
				log.Printf("  - Deleting thread ID: %d", thread.Idforumthread)
			}
			if !c.dryRun {
				if err := queries.AdminDeleteForumThread(ctx, thread.Idforumthread); err != nil {
					log.Printf("  - error deleting thread %d: %v", thread.Idforumthread, err)
				}
			}
		}
	}

	log.Println("Cleanup of empty private forum threads complete.")
	return nil
}
