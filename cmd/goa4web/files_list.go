package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	adminhandlers "github.com/arran4/goa4web/handlers/admin"
)

// filesListCmd implements "files list".
type filesListCmd struct {
	*filesCmd
	fs        *flag.FlagSet
	path      string
	olderThan time.Duration
	jsonOut   bool
}

func parseFilesListCmd(parent *filesCmd, args []string) (*filesListCmd, error) {
	c := &filesListCmd{filesCmd: parent}
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.path, "path", "", "path under the image upload directory")
		fs.DurationVar(&c.olderThan, "older-than", 0, "only include files older than this duration")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

type filesListEntry struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Size     int64      `json:"size"`
	IsDir    bool       `json:"is_dir"`
	Username string     `json:"username,omitempty"`
	Board    string     `json:"board,omitempty"`
	Posted   *time.Time `json:"posted,omitempty"`
	ModTime  *time.Time `json:"mod_time,omitempty"`
}

type filesListSummary struct {
	Entries int   `json:"entries"`
	Files   int   `json:"files"`
	Dirs    int   `json:"dirs"`
	Bytes   int64 `json:"bytes"`
}

type filesListOutput struct {
	Path    string           `json:"path"`
	Entries []filesListEntry `json:"entries"`
	Summary filesListSummary `json:"summary"`
}

func (c *filesListCmd) Run() error {
	queries, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	listing, err := adminhandlers.BuildImageFilesListing(c.rootCmd.Context(), queries, c.rootCmd.cfg.ImageUploadDir, c.path, "", nil, 0)
	if err != nil {
		return err
	}

	filtered, summary := filterImageFiles(listing.Entries, c.olderThan, false)
	output := filesListOutput{
		Path:    listing.Path,
		Entries: filtered,
		Summary: summary,
	}

	if c.jsonOut {
		b, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Path\tSize\tModified\tUsername\tBoard\tPosted\tType")
	for _, entry := range output.Entries {
		modTime := "-"
		if entry.ModTime != nil && !entry.ModTime.IsZero() {
			modTime = entry.ModTime.Format(time.RFC3339)
		}
		posted := "-"
		if entry.Posted != nil && !entry.Posted.IsZero() {
			posted = entry.Posted.Format(time.RFC3339)
		}
		size := "-"
		entryType := "file"
		if entry.IsDir {
			entryType = "dir"
		} else {
			size = fmt.Sprintf("%d", entry.Size)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			entry.Path,
			size,
			modTime,
			blankIfEmpty(entry.Username),
			blankIfEmpty(entry.Board),
			posted,
			entryType,
		)
	}
	w.Flush()
	fmt.Printf("\nSummary: entries=%d files=%d dirs=%d bytes=%d\n",
		output.Summary.Entries,
		output.Summary.Files,
		output.Summary.Dirs,
		output.Summary.Bytes,
	)
	return nil
}

func (c *filesListCmd) Usage() {
	executeUsage(c.fs.Output(), "files_list_usage.txt", c)
}

func (c *filesListCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*filesListCmd)(nil)

func filterImageFiles(entries []adminhandlers.ImageFileEntry, olderThan time.Duration, onlyFiles bool) ([]filesListEntry, filesListSummary) {
	now := time.Now()
	var filtered []filesListEntry
	var summary filesListSummary
	for _, entry := range entries {
		if entry.IsDir && onlyFiles {
			continue
		}
		if olderThan > 0 && !entry.IsDir {
			if entry.ModTime.IsZero() {
				continue
			}
			if now.Sub(entry.ModTime) < olderThan {
				continue
			}
		} else if olderThan > 0 && entry.IsDir {
			continue
		}
		summary.Entries++
		if entry.IsDir {
			summary.Dirs++
		} else {
			summary.Files++
			summary.Bytes += entry.Size
		}
		filtered = append(filtered, filesListEntry{
			Name:     entry.Name,
			Path:     entry.Path,
			Size:     entry.Size,
			IsDir:    entry.IsDir,
			Username: entry.Username,
			Board:    entry.Board,
			Posted:   timePtr(entry.Posted),
			ModTime:  timePtr(entry.ModTime),
		})
	}
	return filtered, summary
}

func blankIfEmpty(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
