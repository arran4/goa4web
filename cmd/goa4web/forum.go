package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/arran4/goa4web/internal/db"
)

const (
	// forumListPageSize is the default page size for forum maintenance listings.
	forumListPageSize = 200
)

func isPublicForumHandler(handler string) bool {
	return handler == ""
}

// forumCmd handles forum utilities.
type forumCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseForumCmd(parent *rootCmd, args []string) (*forumCmd, error) {
	c := &forumCmd{rootCmd: parent}
	c.fs = newFlagSet("forum")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *forumCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing forum command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "clean-empty":
		cmd, err := parseForumCleanEmptyCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("clean-empty: %w", err)
		}
		return cmd.Run()
	case "clean-empty-threads":
		cmd, err := parseForumCleanEmptyThreadsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("clean-empty-threads: %w", err)
		}
		return cmd.Run()
	case "topic":
		cmd, err := parseForumTopicCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("topic: %w", err)
		}
		return cmd.Run()
	case "thread":
		cmd, err := parseForumThreadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("thread: %w", err)
		}
		return cmd.Run()
	case "comment":
		cmd, err := parseForumCommentCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("comment: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown forum command %q", args[0])
	}
}

// Usage prints command usage information.
func (c *forumCmd) Usage() {
	executeUsage(c.fs.Output(), "forum_usage.txt", c)
}

func (c *forumCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

// forumCleanEmptyCmd implements "forum clean-empty".
type forumCleanEmptyCmd struct {
	*forumCmd
	fs       *flag.FlagSet
	dryRun   bool
	verbose  bool
	pageSize int
}

func parseForumCleanEmptyCmd(parent *forumCmd, args []string) (*forumCleanEmptyCmd, error) {
	c := &forumCleanEmptyCmd{forumCmd: parent}
	c.fs = newFlagSet("clean-empty")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "Dry run")
	c.fs.BoolVar(&c.verbose, "verbose", false, "Verbose output")
	c.fs.IntVar(&c.pageSize, "page-size", forumListPageSize, "Number of topics to inspect per page")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *forumCleanEmptyCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	queries := db.New(conn)
	ctx := context.Background()

	if c.dryRun {
		log.Println("Dry run mode enabled. No changes will be made.")
	}

	topicCounts, err := queries.AdminForumTopicThreadCounts(ctx)
	if err != nil {
		return fmt.Errorf("getting topic thread counts: %w", err)
	}
	threadsByTopic := make(map[int32]int64, len(topicCounts))
	for _, tc := range topicCounts {
		if isPublicForumHandler(tc.Handler) {
			threadsByTopic[tc.Idforumtopic] = tc.Threads
		}
	}

	offset := 0
	for {
		topics, err := queries.AdminListForumTopics(ctx, db.AdminListForumTopicsParams{
			Limit:  int32(c.pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("getting forum topics: %w", err)
		}
		if len(topics) == 0 {
			break
		}

		for _, topic := range topics {
			if !isPublicForumHandler(topic.Handler) {
				continue
			}
			threads := threadsByTopic[topic.Idforumtopic]
			if threads == 0 {
				if c.verbose {
					log.Printf("Found empty forum topic: %s (ID: %d)", topic.Title.String, topic.Idforumtopic)
				}

				grants, err := queries.AdminListForumTopicGrantsByTopicID(ctx, sql.NullInt32{Int32: topic.Idforumtopic, Valid: true})
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

		offset += c.pageSize
	}

	log.Println("Cleanup of empty forum topics complete.")
	return nil
}

// forumCleanEmptyThreadsCmd implements "forum clean-empty-threads".
type forumCleanEmptyThreadsCmd struct {
	*forumCmd
	fs       *flag.FlagSet
	dryRun   bool
	verbose  bool
	pageSize int
}

func parseForumCleanEmptyThreadsCmd(parent *forumCmd, args []string) (*forumCleanEmptyThreadsCmd, error) {
	c := &forumCleanEmptyThreadsCmd{forumCmd: parent}
	c.fs = newFlagSet("clean-empty-threads")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "Dry run")
	c.fs.BoolVar(&c.verbose, "verbose", false, "Verbose output")
	c.fs.IntVar(&c.pageSize, "page-size", forumListPageSize, "Number of threads to inspect per page")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *forumCleanEmptyThreadsCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	queries := db.New(conn)
	ctx := context.Background()

	if c.dryRun {
		log.Println("Dry run mode enabled. No changes will be made.")
	}

	offset := 0
	for {
		threads, err := queries.AdminListForumThreads(ctx, db.AdminListForumThreadsParams{
			Limit:  int32(c.pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("getting forum threads: %w", err)
		}
		if len(threads) == 0 {
			break
		}

		for _, thread := range threads {
			if !isPublicForumHandler(thread.TopicHandler) {
				continue
			}
			if !thread.PostCount.Valid || thread.PostCount.Int32 == 0 {
				if c.verbose {
					log.Printf("Found empty forum thread: (ID: %d)", thread.Idforumthread)
				}

				grants, err := queries.AdminListForumThreadGrantsByThreadID(ctx, sql.NullInt32{Int32: thread.Idforumthread, Valid: true})
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

		offset += c.pageSize
	}

	log.Println("Cleanup of empty forum threads complete.")
	return nil
}

var _ usageData = (*forumCmd)(nil)

func isFlagPassed(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
