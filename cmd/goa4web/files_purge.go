package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/handlers/imagebbs"
)

// filesPurgeCmd implements "files purge".
type filesPurgeCmd struct {
	*filesCmd
	fs        *flag.FlagSet
	path      string
	olderThan time.Duration
	dryRun    bool
	jsonOut   bool
}

func parseFilesPurgeCmd(parent *filesCmd, args []string) (*filesPurgeCmd, error) {
	c := &filesPurgeCmd{filesCmd: parent}
	fs, _, err := parseFlags("purge", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.path, "path", "", "path under the image upload directory")
		fs.DurationVar(&c.olderThan, "older-than", 0, "only purge files older than this duration")
		fs.BoolVar(&c.dryRun, "dry-run", false, "preview deletions without removing files")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

type filesPurgeEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type filesPurgeSummary struct {
	Candidates int   `json:"candidates"`
	Deleted    int   `json:"deleted"`
	Errors     int   `json:"errors"`
	Bytes      int64 `json:"bytes"`
	DryRun     bool  `json:"dry_run"`
}

type filesPurgeOutput struct {
	Path    string            `json:"path"`
	Entries []filesPurgeEntry `json:"entries"`
	Summary filesPurgeSummary `json:"summary"`
}

func (c *filesPurgeCmd) Run() error {
	queries, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	listing, err := adminhandlers.BuildImageFilesListing(c.rootCmd.Context(), queries, c.rootCmd.cfg.ImageUploadDir, c.path, "", nil, 0)
	if err != nil {
		return err
	}

	filtered, summary := filterImageFiles(listing.Entries, c.olderThan, true)
	purgeOutput := filesPurgeOutput{
		Path: listing.Path,
		Summary: filesPurgeSummary{
			Candidates: summary.Files,
			DryRun:     c.dryRun,
		},
	}

	base := filepath.Join(c.rootCmd.cfg.ImageUploadDir, imagebbs.ImagebbsUploadPrefix)
	for _, entry := range filtered {
		result := filesPurgeEntry{
			Path: entry.Path,
			Size: entry.Size,
		}
		if c.dryRun {
			result.Status = "dry-run"
			purgeOutput.Summary.Deleted++
			purgeOutput.Summary.Bytes += entry.Size
			purgeOutput.Entries = append(purgeOutput.Entries, result)
			continue
		}
		relPath := strings.TrimPrefix(entry.Path, string(filepath.Separator))
		target := filepath.Join(base, relPath)
		if err := os.Remove(target); err != nil {
			result.Status = "error"
			result.Error = err.Error()
			purgeOutput.Summary.Errors++
		} else {
			result.Status = "deleted"
			purgeOutput.Summary.Deleted++
			purgeOutput.Summary.Bytes += entry.Size
		}
		purgeOutput.Entries = append(purgeOutput.Entries, result)
	}

	if c.jsonOut {
		b, _ := json.MarshalIndent(purgeOutput, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Path\tSize\tStatus\tError")
	for _, entry := range purgeOutput.Entries {
		errMsg := "-"
		if entry.Error != "" {
			errMsg = entry.Error
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", entry.Path, entry.Size, entry.Status, errMsg)
	}
	w.Flush()
	fmt.Printf("\nSummary: candidates=%d deleted=%d errors=%d bytes=%d dry-run=%t\n",
		purgeOutput.Summary.Candidates,
		purgeOutput.Summary.Deleted,
		purgeOutput.Summary.Errors,
		purgeOutput.Summary.Bytes,
		purgeOutput.Summary.DryRun,
	)
	return nil
}

func (c *filesPurgeCmd) Usage() {
	executeUsage(c.fs.Output(), "files_purge_usage.txt", c)
}

func (c *filesPurgeCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*filesPurgeCmd)(nil)
