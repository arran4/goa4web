package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
)

// defaultJobsListLimit caps the number of durable jobs returned by `goa4web jobs list`.
const defaultJobsListLimit = 100

// jobsCmd implements durable background job inspection and operator actions.
type jobsCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseJobsCmd(parent *rootCmd, args []string) (*jobsCmd, error) {
	c := &jobsCmd{rootCmd: parent}
	c.fs = newFlagSet("jobs")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *jobsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing jobs subcommand")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		return c.runList(args[1:])
	case "show":
		return c.runShow(args[1:])
	case "enqueue-url":
		return c.runEnqueueURL(args[1:])
	case "retry":
		return c.runRetry(args[1:])
	case "cancel":
		return c.runCancel(args[1:])
	case "purge":
		return c.runPurge(args[1:])
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown jobs subcommand %q", args[0])
	}
}

func (c *jobsCmd) openService() (*jobs.Service, func(), error) {
	sdb, err := c.getDB()
	if err != nil {
		return nil, nil, err
	}
	q := db.NewForDriver(sdb, c.cfg.DBDriver)
	svc := jobs.NewConfiguredService(q)
	return svc, func() { closeDB(sdb) }, nil
}

func (c *jobsCmd) runList(args []string) error {
	var statusFilter, typeFilter string
	var limit, offset int
	var jsonOut bool
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.StringVar(&statusFilter, "status", "", "Filter by status (pending, running, completed, failed, cancelled)")
		fs.StringVar(&typeFilter, "type", "", "Filter by job type")
		fs.IntVar(&limit, "limit", defaultJobsListLimit, "Maximum jobs to return")
		fs.IntVar(&offset, "offset", 0, "Offset")
		fs.BoolVar(&jsonOut, "json", false, "Output JSON")
	})
	if err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	svc, cleanup, err := c.openService()
	if err != nil {
		return err
	}
	defer cleanup()

	rows, err := svc.AdminList(context.Background(), statusFilter, typeFilter, limit, offset)
	if err != nil {
		return err
	}
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tTYPE\tSTATUS\tATTEMPTS\tNEXT RUN\tKEY\tERROR")
	for _, r := range rows {
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%d/%d\t%s\t%s\t%s\n", r.ID, r.JobType, r.Status, r.Attempts, r.MaxAttempts, r.RunAfter.UTC().Format(time.RFC3339), r.TargetKey, r.LastError.String)
	}
	return tw.Flush()
}

func (c *jobsCmd) runShow(args []string) error {
	var jsonOut bool
	fs, _, err := parseFlags("show", args, func(fs *flag.FlagSet) {
		fs.BoolVar(&jsonOut, "json", false, "Output JSON")
	})
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: goa4web jobs show [--json] <id>")
	}
	id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
	if err != nil || id <= 0 {
		return fmt.Errorf("invalid job id %q", fs.Arg(0))
	}
	svc, cleanup, err := c.openService()
	if err != nil {
		return err
	}
	defer cleanup()

	details, err := svc.AdminGet(context.Background(), id)
	if err != nil {
		return err
	}
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(details)
	}
	job := details.Job
	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "ID:\t%d\n", job.ID)
	_, _ = fmt.Fprintf(tw, "Type:\t%s\n", job.JobType)
	_, _ = fmt.Fprintf(tw, "Status:\t%s\n", job.Status)
	_, _ = fmt.Fprintf(tw, "Target:\t%s:%d:%s\n", job.TargetType, job.TargetID, job.TargetKey)
	_, _ = fmt.Fprintf(tw, "TargetID:\t%d\n", job.TargetID)
	_, _ = fmt.Fprintf(tw, "Revision:\t%d\n", job.TargetRevision)
	_, _ = fmt.Fprintf(tw, "NeedsRerun:\t%t\n", job.NeedsRerun)
	_, _ = fmt.Fprintf(tw, "Attempts:\t%d/%d\n", job.Attempts, job.MaxAttempts)
	_, _ = fmt.Fprintf(tw, "NextRun:\t%s\n", job.RunAfter.UTC().Format(time.RFC3339))
	if job.LeaseOwner.Valid {
		_, _ = fmt.Fprintf(tw, "LeaseOwner:\t%s\n", job.LeaseOwner.String)
	}
	if job.LeaseExpiresAt.Valid {
		_, _ = fmt.Fprintf(tw, "LeaseExpires:\t%s\n", job.LeaseExpiresAt.Time.UTC().Format(time.RFC3339))
	}
	if job.LastError.Valid {
		_, _ = fmt.Fprintf(tw, "LastError:\t%s\n", job.LastError.String)
	}
	if job.CreatedByUserID.Valid {
		_, _ = fmt.Fprintf(tw, "CreatedBy:\t%d\n", job.CreatedByUserID.Int32)
	}
	_, _ = fmt.Fprintf(tw, "CreatedAt:\t%s\n", job.CreatedAt.UTC().Format(time.RFC3339))
	_, _ = fmt.Fprintf(tw, "UpdatedAt:\t%s\n", job.UpdatedAt.UTC().Format(time.RFC3339))
	if job.CompletedAt.Valid {
		_, _ = fmt.Fprintf(tw, "CompletedAt:\t%s\n", job.CompletedAt.Time.UTC().Format(time.RFC3339))
	}
	if details.Cooldown != nil {
		_, _ = fmt.Fprintf(tw, "CooldownUntil:\t%s\n", details.Cooldown.CooldownUntil.UTC().Format(time.RFC3339))
		if details.Cooldown.LastNotifiedAt.Valid {
			_, _ = fmt.Fprintf(tw, "LastNotifiedAt:\t%s\n", details.Cooldown.LastNotifiedAt.Time.UTC().Format(time.RFC3339))
		}
	}
	return tw.Flush()
}

func (c *jobsCmd) runEnqueueURL(args []string) error {
	var force bool
	fs, _, err := parseFlags("enqueue-url", args, func(fs *flag.FlagSet) {
		fs.BoolVar(&force, "force", false, "Bypass active failure cooldown")
	})
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: goa4web jobs enqueue-url [--force] <url>")
	}
	canonical := common.ValidateEnqueueExternalURL(fs.Arg(0))
	if canonical == "" {
		return fmt.Errorf("invalid or disallowed external URL %q", fs.Arg(0))
	}
	svc, cleanup, err := c.openService()
	if err != nil {
		return err
	}
	defer cleanup()

	linkID, created, err := svc.AdminEnqueueExternalURL(context.Background(), canonical, sql.NullInt32{}, force)
	if err != nil {
		return err
	}
	if created {
		fmt.Printf("Enqueued external_link job for %s (external_link #%d)\n", canonical, linkID)
	} else {
		fmt.Printf("Coalesced active external_link job for %s (external_link #%d)\n", canonical, linkID)
	}
	return nil
}

func (c *jobsCmd) runRetry(args []string) error {
	fs, _, err := parseFlags("retry", args, nil)
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: goa4web jobs retry <id>")
	}
	id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
	if err != nil || id <= 0 {
		return fmt.Errorf("invalid job id %q", fs.Arg(0))
	}
	svc, cleanup, err := c.openService()
	if err != nil {
		return err
	}
	defer cleanup()
	ok, err := svc.AdminRetry(context.Background(), id)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("job %d not in a terminal state", id)
	}
	fmt.Printf("Re-queued durable job %d\n", id)
	return nil
}

func (c *jobsCmd) runCancel(args []string) error {
	fs, _, err := parseFlags("cancel", args, nil)
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: goa4web jobs cancel <id>")
	}
	id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
	if err != nil || id <= 0 {
		return fmt.Errorf("invalid job id %q", fs.Arg(0))
	}
	svc, cleanup, err := c.openService()
	if err != nil {
		return err
	}
	defer cleanup()
	ok, err := svc.AdminCancel(context.Background(), id)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("job %d not pending or running", id)
	}
	fmt.Printf("Cancelled durable job %d\n", id)
	return nil
}

func (c *jobsCmd) runPurge(args []string) error {
	var statusFilter string
	fs, _, err := parseFlags("purge", args, func(fs *flag.FlagSet) {
		fs.StringVar(&statusFilter, "status", "", "Terminal status to purge (completed, failed, cancelled; empty = all terminal)")
	})
	if err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	svc, cleanup, err := c.openService()
	if err != nil {
		return err
	}
	defer cleanup()
	n, err := svc.AdminPurge(context.Background(), statusFilter)
	if err != nil {
		return err
	}
	fmt.Printf("Purged %d terminal durable jobs\n", n)
	return nil
}

// Usage prints command usage information for `goa4web jobs`.
func (c *jobsCmd) Usage() {
	fmt.Fprintf(os.Stderr, `Usage: goa4web jobs <subcommand> [flags]

Subcommands:
  list         List durable background jobs
  show         Show metadata and cooldown state for one job
  enqueue-url  Enqueue an external URL metadata fetch job
  retry        Retry a failed/cancelled/completed job (clears cooldown)
  cancel       Cancel a pending/running job
  purge        Delete terminal job rows while preserving cooldowns
`)
}
