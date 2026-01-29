package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// requestsListCmd implements "requests list".
type requestsListCmd struct {
	*requestsCmd
	fs       *flag.FlagSet
	status   string
	kind     string
	json     bool
	offset   int
	pageSize int
}

type requestJSON struct {
	ID             int32   `json:"id"`
	UserID         int32   `json:"user_id"`
	ChangeTable    string  `json:"change_table"`
	ChangeField    string  `json:"change_field"`
	ChangeRowID    int32   `json:"change_row_id"`
	ChangeValue    *string `json:"change_value"`
	ContactOptions *string `json:"contact_options"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	ActedAt        *string `json:"acted_at"`
}

const (
	// requestStatusPending filters requests that are awaiting action.
	requestStatusPending = "pending"
	// requestStatusArchived filters requests that are no longer pending.
	requestStatusArchived = "archived"
)

func parseRequestsListCmd(parent *requestsCmd, args []string) (*requestsListCmd, error) {
	c := &requestsListCmd{requestsCmd: parent}
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.status, "status", requestStatusPending, "filter by request status (pending, archived, accepted, rejected, query)")
		fs.StringVar(&c.kind, "type", "", "filter by request change table")
		fs.BoolVar(&c.json, "json", false, "machine-readable JSON output")
		fs.IntVar(&c.offset, "offset", 0, "pagination offset")
		fs.IntVar(&c.pageSize, "page-size", 0, "page size (0 uses the configured default)")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

// requestListItem is a JSON representation of a request queue entry.
type requestListItem struct {
	ID             int32   `json:"id"`
	UserID         int32   `json:"user_id"`
	ChangeTable    string  `json:"change_table"`
	ChangeField    string  `json:"change_field"`
	ChangeRowID    int32   `json:"change_row_id"`
	ChangeValue    *string `json:"change_value"`
	ContactOptions *string `json:"contact_options"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	ActedAt        *string `json:"acted_at"`
}

type requestsListOutput struct {
	Status    string        `json:"status"`
	Offset    int           `json:"offset"`
	PageSize  int           `json:"page_size"`
	Total     int           `json:"total"`
	HasMore   bool          `json:"has_more"`
	Requests  []requestJSON `json:"requests"`
	Requested int           `json:"requested"`
}

func (c *requestsListCmd) Usage() {
	executeUsage(c.fs.Output(), "requests_list_usage.txt", c)
}

func (c *requestsListCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*requestsListCmd)(nil)

func (c *requestsListCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)

	// Determine status filter
	status := strings.ToLower(strings.TrimSpace(c.status))

	// Load requests based on status
	var rows []*db.AdminRequestQueue
	switch status {
	case requestStatusPending:
		rows, err = queries.AdminListPendingRequests(ctx)
	case requestStatusArchived:
		rows, err = queries.AdminListArchivedRequests(ctx)
	default:
		// Try generic lookup if supported, otherwise error or fallback
		// The original code had a generic AdminListRequestQueueByStatus call in the second implementation.
		rows, err = queries.AdminListRequestQueueByStatus(ctx, status)
	}
	if err != nil {
		return fmt.Errorf("list requests: %w", err)
	}

	// Filter by kind/type if specified
	if c.kind != "" {
		filtered := make([]*db.AdminRequestQueue, 0, len(rows))
		for _, row := range rows {
			if row.ChangeTable == c.kind {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}

	// Determine page size
	pageSize := c.pageSize
	if pageSize <= 0 {
		cfg, cfgErr := c.rootCmd.RuntimeConfig()
		if cfgErr == nil {
			pageSize = cfg.PageSizeDefault
			if pageSize < cfg.PageSizeMin {
				pageSize = cfg.PageSizeMin
			}
			if pageSize > cfg.PageSizeMax {
				pageSize = cfg.PageSizeMax
			}
		} else {
			// Fallback if config fails
			pageSize = 20
		}
	}
	if pageSize < 1 {
		pageSize = 1
	}
	if c.offset < 0 {
		c.offset = 0
	}

	// Apply pagination
	total := len(rows)
	start := c.offset
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	hasMore := end < total
	selected := rows[start:end]

	if c.json {
		// Output JSON structure with metadata
		items := make([]requestJSON, 0, len(selected))
		for _, row := range selected {
			items = append(items, requestToJSON(row))
		}

		payload := requestsListOutput{
			Status:    status,
			Offset:    c.offset,
			PageSize:  pageSize,
			Total:     total,
			HasMore:   hasMore,
			Requests:  items,
			Requested: len(items),
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Output Table
	return c.printTable(selected)
}

func requestToJSON(req *db.AdminRequestQueue) requestJSON {
	return requestJSON{
		ID:             req.ID,
		UserID:         req.UsersIdusers,
		ChangeTable:    req.ChangeTable,
		ChangeField:    req.ChangeField,
		ChangeRowID:    req.ChangeRowID,
		ChangeValue:    optionalString(req.ChangeValue),
		ContactOptions: optionalString(req.ContactOptions),
		Status:         req.Status,
		CreatedAt:      req.CreatedAt.Format(time.RFC3339),
		ActedAt:        optionalTime(req.ActedAt),
	}
}

func optionalString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	val := value.String
	return &val
}

func optionalTime(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	val := value.Time.Format(time.RFC3339)
	return &val
}

func (c *requestsListCmd) printTable(rows []*db.AdminRequestQueue) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tUserID\tTable\tField\tRowID\tValue\tContact\tStatus\tCreated\tActed")
	for _, row := range rows {
		created := row.CreatedAt.Format(time.RFC3339)
		acted := "-"
		if row.ActedAt.Valid {
			acted = row.ActedAt.Time.Format(time.RFC3339)
		}
		fmt.Fprintf(
			w,
			"%d\t%d\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.UsersIdusers,
			row.ChangeTable,
			row.ChangeField,
			row.ChangeRowID,
			reqNullStringValue(row.ChangeValue),
			reqNullStringValue(row.ContactOptions),
			row.Status,
			created,
			acted,
		)
	}
	return w.Flush()
}

func reqNullStringValue(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func reqNullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	value := ns.String
	return &value
}
