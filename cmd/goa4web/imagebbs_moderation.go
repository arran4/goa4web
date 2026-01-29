package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

// imagebbsModerationCmd handles ImageBBS moderation subcommands.
type imagebbsModerationCmd struct {
	*imagebbsCmd
	fs *flag.FlagSet
}

func parseImagebbsModerationCmd(parent *imagebbsCmd, args []string) (*imagebbsModerationCmd, error) {
	c := &imagebbsModerationCmd{imagebbsCmd: parent}
	c.fs = newFlagSet("moderation")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *imagebbsModerationCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing moderation command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "approve":
		cmd, err := parseImagebbsModerationApproveCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("approve: %w", err)
		}
		return cmd.Run()
	case "reject":
		cmd, err := parseImagebbsModerationRejectCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("reject: %w", err)
		}
		return cmd.Run()
	case "bulk-approve":
		cmd, err := parseImagebbsModerationBulkApproveCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("bulk-approve: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown moderation command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *imagebbsModerationCmd) Usage() {
	executeUsage(c.fs.Output(), "imagebbs_moderation_usage.txt", c)
}

func (c *imagebbsModerationCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*imagebbsModerationCmd)(nil)

// imagebbsModerationApproveCmd approves a single image post.
type imagebbsModerationApproveCmd struct {
	*imagebbsModerationCmd
	fs *flag.FlagSet
	ID int
}

func parseImagebbsModerationApproveCmd(parent *imagebbsModerationCmd, args []string) (*imagebbsModerationApproveCmd, error) {
	c := &imagebbsModerationApproveCmd{imagebbsModerationCmd: parent}
	c.fs = newFlagSet("approve")
	c.fs.IntVar(&c.ID, "id", 0, "image post id")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	switch remaining := c.fs.Args(); len(remaining) {
	case 0:
	case 1:
		if c.ID != 0 {
			return nil, fmt.Errorf("unexpected arguments: %v", remaining)
		}
		id, err := strconv.Atoi(remaining[0])
		if err != nil {
			return nil, fmt.Errorf("invalid id %q", remaining[0])
		}
		c.ID = id
	default:
		return nil, fmt.Errorf("unexpected arguments: %v", remaining)
	}
	return c, nil
}

func (c *imagebbsModerationApproveCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	q, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("querier: %w", err)
	}
	ctx := c.rootCmd.Context()
	actionErr := q.AdminApproveImagePost(ctx, int32(c.ID))
	result := imagebbsModerationResult{ID: int32(c.ID), Action: "approve", Status: "approved", Err: actionErr}
	if actionErr != nil {
		result.Status = "error"
	}
	if err := printImagebbsModerationResults(c.fs.Output(), []imagebbsModerationResult{result}); err != nil {
		return fmt.Errorf("print results: %w", err)
	}
	if actionErr != nil {
		return fmt.Errorf("approve image post %d: %w", c.ID, actionErr)
	}
	return nil
}

func (c *imagebbsModerationApproveCmd) Usage() {
	executeUsage(c.fs.Output(), "imagebbs_moderation_approve_usage.txt", c)
}

func (c *imagebbsModerationApproveCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*imagebbsModerationApproveCmd)(nil)

// imagebbsModerationRejectCmd rejects a single image post.
type imagebbsModerationRejectCmd struct {
	*imagebbsModerationCmd
	fs *flag.FlagSet
	ID int
}

func parseImagebbsModerationRejectCmd(parent *imagebbsModerationCmd, args []string) (*imagebbsModerationRejectCmd, error) {
	c := &imagebbsModerationRejectCmd{imagebbsModerationCmd: parent}
	c.fs = newFlagSet("reject")
	c.fs.IntVar(&c.ID, "id", 0, "image post id")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	switch remaining := c.fs.Args(); len(remaining) {
	case 0:
	case 1:
		if c.ID != 0 {
			return nil, fmt.Errorf("unexpected arguments: %v", remaining)
		}
		id, err := strconv.Atoi(remaining[0])
		if err != nil {
			return nil, fmt.Errorf("invalid id %q", remaining[0])
		}
		c.ID = id
	default:
		return nil, fmt.Errorf("unexpected arguments: %v", remaining)
	}
	return c, nil
}

func (c *imagebbsModerationRejectCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	q, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("querier: %w", err)
	}
	ctx := c.rootCmd.Context()
	actionErr := q.AdminDeleteImagePost(ctx, int32(c.ID))
	result := imagebbsModerationResult{ID: int32(c.ID), Action: "reject", Status: "rejected", Err: actionErr}
	if actionErr != nil {
		result.Status = "error"
	}
	if err := printImagebbsModerationResults(c.fs.Output(), []imagebbsModerationResult{result}); err != nil {
		return fmt.Errorf("print results: %w", err)
	}
	if actionErr != nil {
		return fmt.Errorf("reject image post %d: %w", c.ID, actionErr)
	}
	return nil
}

func (c *imagebbsModerationRejectCmd) Usage() {
	executeUsage(c.fs.Output(), "imagebbs_moderation_reject_usage.txt", c)
}

func (c *imagebbsModerationRejectCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*imagebbsModerationRejectCmd)(nil)

// imagebbsModerationBulkApproveCmd approves multiple image posts.
type imagebbsModerationBulkApproveCmd struct {
	*imagebbsModerationCmd
	fs       *flag.FlagSet
	FromFile string
	args     []string
}

func parseImagebbsModerationBulkApproveCmd(parent *imagebbsModerationCmd, args []string) (*imagebbsModerationBulkApproveCmd, error) {
	c := &imagebbsModerationBulkApproveCmd{imagebbsModerationCmd: parent}
	fs, remaining, err := parseFlags("bulk-approve", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.FromFile, "from-file", "", "path to file containing image post ids (use '-' for stdin)")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.fs.Usage = c.Usage
	c.args = remaining
	return c, nil
}

func (c *imagebbsModerationBulkApproveCmd) Run() error {
	ids, err := collectImagePostIDs(c.FromFile, c.args)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return fmt.Errorf("at least one id required")
	}
	q, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("querier: %w", err)
	}
	ctx := c.rootCmd.Context()
	results := make([]imagebbsModerationResult, 0, len(ids))
	failed := 0
	for _, id := range ids {
		actionErr := q.AdminApproveImagePost(ctx, id)
		result := imagebbsModerationResult{ID: id, Action: "approve", Status: "approved", Err: actionErr}
		if actionErr != nil {
			result.Status = "error"
			failed++
		}
		results = append(results, result)
	}
	if err := printImagebbsModerationResults(c.fs.Output(), results); err != nil {
		return fmt.Errorf("print results: %w", err)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d approvals failed", failed, len(ids))
	}
	return nil
}

func (c *imagebbsModerationBulkApproveCmd) Usage() {
	executeUsage(c.fs.Output(), "imagebbs_moderation_bulk_approve_usage.txt", c)
}

func (c *imagebbsModerationBulkApproveCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*imagebbsModerationBulkApproveCmd)(nil)

type imagebbsModerationResult struct {
	ID     int32
	Action string
	Status string
	Err    error
}

func collectImagePostIDs(fromFile string, args []string) ([]int32, error) {
	var ids []int32
	if fromFile != "" {
		reader, closeFn, err := openIDsReader(fromFile)
		if err != nil {
			return nil, err
		}
		fileIDs, err := parseImagePostIDs(reader)
		if closeFn != nil {
			if closeErr := closeFn(); closeErr != nil && err == nil {
				err = closeErr
			}
		}
		if err != nil {
			return nil, err
		}
		ids = append(ids, fileIDs...)
	}
	if len(args) > 0 {
		argIDs, err := parseImagePostIDs(strings.NewReader(strings.Join(args, " ")))
		if err != nil {
			return nil, err
		}
		ids = append(ids, argIDs...)
	}
	return ids, nil
}

func openIDsReader(path string) (io.Reader, func() error, error) {
	if path == "-" {
		return os.Stdin, nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open from-file: %w", err)
	}
	return file, file.Close, nil
}

func parseImagePostIDs(r io.Reader) ([]int32, error) {
	scanner := bufio.NewScanner(r)
	var ids []int32
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.FieldsFunc(line, func(r rune) bool {
			switch r {
			case ' ', '\t', ',', ';':
				return true
			default:
				return false
			}
		})
		for _, field := range fields {
			if field == "" {
				continue
			}
			id, err := strconv.Atoi(field)
			if err != nil {
				return nil, fmt.Errorf("invalid id %q", field)
			}
			ids = append(ids, int32(id))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read ids: %w", err)
	}
	return ids, nil
}

func printImagebbsModerationResults(out io.Writer, results []imagebbsModerationResult) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAction\tStatus\tError")
	for _, result := range results {
		errText := "-"
		if result.Err != nil {
			errText = result.Err.Error()
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", result.ID, result.Action, result.Status, errText)
	}
	return w.Flush()
}
