package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/internal/stats"
)

const (
	// serverStatsFormatTable is the default output format for server stats.
	serverStatsFormatTable = "table"
	// serverStatsFormatJSON is the JSON output format for server stats.
	serverStatsFormatJSON = "json"
)

// serverStatsCmd implements "server stats".
type serverStatsCmd struct {
	*serverCmd
	fs      *flag.FlagSet
	Format  string
	StartAt string
	EndAt   string
}

func parseServerStatsCmd(parent *serverCmd, args []string) (*serverStatsCmd, error) {
	c := &serverStatsCmd{serverCmd: parent}
	c.fs = newFlagSet("stats")
	c.fs.StringVar(&c.Format, "format", serverStatsFormatTable, "output format (table, json)")
	c.fs.StringVar(&c.StartAt, "start", "", "optional start time (RFC3339)")
	c.fs.StringVar(&c.EndAt, "end", "", "optional end time (RFC3339)")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *serverStatsCmd) Run() error {
	format := strings.ToLower(strings.TrimSpace(c.Format))
	if format == "" {
		format = serverStatsFormatTable
	}
	if format != serverStatsFormatTable && format != serverStatsFormatJSON {
		return fmt.Errorf("invalid format %q (expected %s or %s)", c.Format, serverStatsFormatTable, serverStatsFormatJSON)
	}

	startAt, err := parseServerStatsTimestamp(c.StartAt)
	if err != nil {
		return err
	}
	endAt, err := parseServerStatsTimestamp(c.EndAt)
	if err != nil {
		return err
	}
	if startAt != nil && endAt != nil && endAt.Before(*startAt) {
		return fmt.Errorf("start time is after end time")
	}

	data := stats.BuildServerStatsData(c.cfg, c.ConfigFile, c.tasksReg, c.dbReg, c.dlqReg, c.emailReg, c.routerReg.Names())
	uptime := data.Uptime.String()
	if stats.StartTime.IsZero() {
		uptime = "unknown"
	}

	if format == serverStatsFormatJSON {
		payload := serverStatsOutput{
			Stats:        data.Stats,
			Uptime:       uptime,
			Registries:   data.Registries,
			ConfigEnv:    data.ConfigEnv,
			ConfigValues: data.ConfigValues,
		}
		if startAt != nil {
			payload.RangeStart = startAt.Format(time.RFC3339)
		}
		if endAt != nil {
			payload.RangeEnd = endAt.Format(time.RFC3339)
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("json output: %w", err)
		}
		fmt.Println(string(b))
		return nil
	}

	if err := renderServerStatsTable(data, uptime, startAt, endAt); err != nil {
		return err
	}
	return nil
}

type serverStatsOutput struct {
	Stats        stats.ServerStatsMetrics    `json:"stats"`
	Uptime       string                      `json:"uptime"`
	Registries   stats.ServerStatsRegistries `json:"registries"`
	ConfigEnv    string                      `json:"config_env,omitempty"`
	ConfigValues map[string]string           `json:"config_values,omitempty"`
	RangeStart   string                      `json:"range_start,omitempty"`
	RangeEnd     string                      `json:"range_end,omitempty"`
}

func renderServerStatsTable(data stats.ServerStatsData, uptime string, startAt *time.Time, endAt *time.Time) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "Metric\tValue")
	_, _ = fmt.Fprintf(w, "Uptime\t%s\n", uptime)
	if startAt != nil {
		_, _ = fmt.Fprintf(w, "Range Start\t%s\n", startAt.Format(time.RFC3339))
	}
	if endAt != nil {
		_, _ = fmt.Fprintf(w, "Range End\t%s\n", endAt.Format(time.RFC3339))
	}
	_, _ = fmt.Fprintf(w, "Goroutines\t%d\n", data.Stats.Goroutines)
	_, _ = fmt.Fprintf(w, "Alloc\t%d\n", data.Stats.Alloc)
	_, _ = fmt.Fprintf(w, "Total Alloc\t%d\n", data.Stats.TotalAlloc)
	_, _ = fmt.Fprintf(w, "System\t%d\n", data.Stats.Sys)
	_, _ = fmt.Fprintf(w, "Heap Alloc\t%d\n", data.Stats.HeapAlloc)
	_, _ = fmt.Fprintf(w, "Heap Sys\t%d\n", data.Stats.HeapSys)
	_, _ = fmt.Fprintf(w, "GC Count\t%d\n", data.Stats.NumGC)
	_, _ = fmt.Fprintf(w, "CPU Cores\t%d\n", data.Stats.NumCPU)
	_, _ = fmt.Fprintf(w, "Architecture\t%s\n", data.Stats.Arch)
	_, _ = fmt.Fprintf(w, "Disk Free\t%d\n", data.Stats.DiskFree)
	_, _ = fmt.Fprintf(w, "RAM Free\t%d\n", data.Stats.RAMFree)
	_, _ = fmt.Fprintf(w, "Tasks\t%s\n", strings.Join(data.Registries.Tasks, ", "))
	_, _ = fmt.Fprintf(w, "Database Drivers\t%s\n", strings.Join(data.Registries.DBDrivers, ", "))
	_, _ = fmt.Fprintf(w, "DLQ Providers\t%s\n", strings.Join(data.Registries.DLQProviders, ", "))
	_, _ = fmt.Fprintf(w, "Email Providers\t%s\n", strings.Join(data.Registries.EmailProviders, ", "))
	_, _ = fmt.Fprintf(w, "Upload Providers\t%s\n", strings.Join(data.Registries.UploadProviders, ", "))
	_, _ = fmt.Fprintf(w, "Router Modules\t%s\n", strings.Join(data.Registries.RouterModules, ", "))
	_ = w.Flush()

	if data.ConfigEnv != "" {
		_, _ = fmt.Fprintln(os.Stdout, "\nConfig (env):")
		_, _ = fmt.Fprint(os.Stdout, data.ConfigEnv)
	}
	if data.ConfigJSON != "" {
		_, _ = fmt.Fprintln(os.Stdout, "\nConfig (json):")
		_, _ = fmt.Fprintln(os.Stdout, data.ConfigJSON)
	}
	return nil
}

func parseServerStatsTimestamp(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp %q (expected RFC3339)", raw)
		}
	}
	return &parsed, nil
}
