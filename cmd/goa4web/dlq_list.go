package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"text/tabwriter"
	"time"

	dirdlq "github.com/arran4/goa4web/internal/dlq/dir"
	filedlq "github.com/arran4/goa4web/internal/dlq/file"
)

// defaultDLQListLimit sets the default maximum number of entries returned per provider.
const defaultDLQListLimit = 100

// dlqListCmd implements "dlq list".
type dlqListCmd struct {
	*dlqCmd
	fs      *flag.FlagSet
	offset  int
	limit   int
	jsonOut bool
}

type dlqListEntry struct {
	ID      int32  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Time    string `json:"time,omitempty"`
	Message string `json:"message"`
}

type dlqListProviderOutput struct {
	Provider string         `json:"provider"`
	Source   string         `json:"source,omitempty"`
	Total    *int64         `json:"total,omitempty"`
	Latest   string         `json:"latest,omitempty"`
	Entries  []dlqListEntry `json:"entries"`
}

type dlqListOutput struct {
	Offset    int                     `json:"offset"`
	Limit     int                     `json:"limit"`
	Providers []dlqListProviderOutput `json:"providers"`
}

func parseDlqListCmd(parent *dlqCmd, args []string) (*dlqListCmd, error) {
	c := &dlqListCmd{dlqCmd: parent}
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.offset, "offset", 0, "number of entries to skip per provider")
		fs.IntVar(&c.limit, "limit", defaultDLQListLimit, "max entries to return per provider")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.fs.Usage = c.Usage
	return c, nil
}

func (c *dlqListCmd) Run() error {
	if c.offset < 0 {
		return fmt.Errorf("offset must be >= 0")
	}
	if c.limit < 0 {
		return fmt.Errorf("limit must be >= 0")
	}
	if c.limit == 0 {
		c.limit = defaultDLQListLimit
	}

	providers, err := c.providers()
	if err != nil {
		return err
	}

	output := dlqListOutput{Offset: c.offset, Limit: c.limit}
	for _, provider := range providers {
		providerOutput, err := c.listProvider(provider)
		if err != nil {
			return err
		}
		output.Providers = append(output.Providers, providerOutput)
	}

	if c.jsonOut {
		b, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		fmt.Fprintln(c.fs.Output(), string(b))
		return nil
	}

	w := tabwriter.NewWriter(c.fs.Output(), 0, 0, 2, ' ', 0)
	for _, provider := range output.Providers {
		fmt.Fprintf(w, "Provider:\t%s\n", provider.Provider)
		if provider.Source != "" {
			fmt.Fprintf(w, "Source:\t%s\n", provider.Source)
		}
		if provider.Total != nil {
			fmt.Fprintf(w, "Total:\t%d\n", *provider.Total)
		}
		if provider.Latest != "" {
			fmt.Fprintf(w, "Latest:\t%s\n", provider.Latest)
		}
		switch provider.Provider {
		case "db":
			fmt.Fprintln(w, "ID\tCreated At\tMessage")
			for _, entry := range provider.Entries {
				fmt.Fprintf(w, "%d\t%s\t%s\n", entry.ID, entry.Time, entry.Message)
			}
		case "file":
			fmt.Fprintln(w, "Time\tMessage")
			for _, entry := range provider.Entries {
				fmt.Fprintf(w, "%s\t%s\n", entry.Time, entry.Message)
			}
		case "dir":
			fmt.Fprintln(w, "Name\tMessage")
			for _, entry := range provider.Entries {
				fmt.Fprintf(w, "%s\t%s\n", entry.Name, entry.Message)
			}
		default:
			fmt.Fprintln(w, "Entry\tMessage")
			for _, entry := range provider.Entries {
				fmt.Fprintf(w, "%s\t%s\n", entry.Name, entry.Message)
			}
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func (c *dlqListCmd) listProvider(provider string) (dlqListProviderOutput, error) {
	switch provider {
	case "db":
		return c.listDB()
	case "file":
		return c.listFile()
	case "dir":
		return c.listDir()
	default:
		return dlqListProviderOutput{}, fmt.Errorf("unsupported dlq provider %q", provider)
	}
}

func (c *dlqListCmd) listDB() (dlqListProviderOutput, error) {
	queries, err := c.rootCmd.Querier()
	if err != nil {
		return dlqListProviderOutput{}, fmt.Errorf("database: %w", err)
	}
	fetchLimit := c.limit + c.offset
	if fetchLimit > math.MaxInt32 {
		return dlqListProviderOutput{}, fmt.Errorf("limit exceeds maximum")
	}
	rows, err := queries.SystemListDeadLetters(c.rootCmd.Context(), int32(fetchLimit))
	if err != nil {
		return dlqListProviderOutput{}, fmt.Errorf("list dead letters: %w", err)
	}
	if c.offset > len(rows) {
		rows = nil
	} else if c.offset > 0 {
		rows = rows[c.offset:]
	}
	entries := make([]dlqListEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, dlqListEntry{
			ID:      row.ID,
			Time:    row.CreatedAt.Format(time.RFC3339),
			Message: row.Message,
		})
	}
	count, err := queries.SystemCountDeadLetters(c.rootCmd.Context())
	if err != nil {
		return dlqListProviderOutput{}, fmt.Errorf("count dead letters: %w", err)
	}
	latestStr := ""
	if latest, err := queries.SystemLatestDeadLetter(c.rootCmd.Context()); err == nil {
		if latestTime, ok := latest.(time.Time); ok {
			latestStr = latestTime.Format(time.RFC3339)
		}
	}
	return dlqListProviderOutput{
		Provider: "db",
		Total:    &count,
		Latest:   latestStr,
		Entries:  entries,
	}, nil
}

func (c *dlqListCmd) listFile() (dlqListProviderOutput, error) {
	path := c.rootCmd.cfg.DLQFile
	if path == "" {
		return dlqListProviderOutput{}, fmt.Errorf("dlq file path not configured")
	}
	fetchLimit := c.limit + c.offset
	if fetchLimit < 0 {
		fetchLimit = 0
	}
	recs, err := filedlq.List(path, fetchLimit)
	if err != nil {
		return dlqListProviderOutput{}, fmt.Errorf("list file dlq: %w", err)
	}
	if c.offset > len(recs) {
		recs = nil
	} else if c.offset > 0 {
		recs = recs[c.offset:]
	}
	entries := make([]dlqListEntry, 0, len(recs))
	for _, rec := range recs {
		entries = append(entries, dlqListEntry{
			Time:    rec.Time.Format(time.RFC3339),
			Message: rec.Message,
		})
	}
	return dlqListProviderOutput{
		Provider: "file",
		Source:   path,
		Entries:  entries,
	}, nil
}

func (c *dlqListCmd) listDir() (dlqListProviderOutput, error) {
	path := c.rootCmd.cfg.DLQFile
	if path == "" {
		return dlqListProviderOutput{}, fmt.Errorf("dlq dir path not configured")
	}
	fetchLimit := c.limit + c.offset
	if fetchLimit < 0 {
		fetchLimit = 0
	}
	recs, err := dirdlq.List(path, fetchLimit)
	if err != nil {
		return dlqListProviderOutput{}, fmt.Errorf("list dir dlq: %w", err)
	}
	if c.offset > len(recs) {
		recs = nil
	} else if c.offset > 0 {
		recs = recs[c.offset:]
	}
	entries := make([]dlqListEntry, 0, len(recs))
	for _, rec := range recs {
		entries = append(entries, dlqListEntry{
			Name:    rec.Name,
			Message: rec.Message,
		})
	}
	return dlqListProviderOutput{
		Provider: "dir",
		Source:   path,
		Entries:  entries,
	}, nil
}

// Usage prints command usage information with examples.
func (c *dlqListCmd) Usage() {
	executeUsage(c.fs.Output(), "dlq_list_usage.txt", c)
}

func (c *dlqListCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*dlqListCmd)(nil)
