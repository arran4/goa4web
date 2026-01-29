package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// requestsListCmd implements "requests list".
type requestsListCmd struct {
	*requestsCmd
	fs       *flag.FlagSet
	status   string
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

type requestsListOutput struct {
	Status    string        `json:"status"`
	Offset    int           `json:"offset"`
	PageSize  int           `json:"page_size"`
	Total     int           `json:"total"`
	HasMore   bool          `json:"has_more"`
	Requests  []requestJSON `json:"requests"`
	Requested int           `json:"requested"`
}

func parseRequestsListCmd(parent *requestsCmd, args []string) (*requestsListCmd, error) {
	c := &requestsListCmd{requestsCmd: parent}
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.status, "status", "pending", "request status (pending or archived)")
		fs.IntVar(&c.offset, "offset", 0, "pagination offset")
		fs.IntVar(&c.pageSize, "page-size", 0, "page size (0 uses the configured default)")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
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

	status := strings.ToLower(strings.TrimSpace(c.status))
	var rows []*db.AdminRequestQueue
	switch status {
	case "pending":
		rows, err = queries.AdminListPendingRequests(ctx)
	case "archived":
		rows, err = queries.AdminListArchivedRequests(ctx)
	default:
		return fmt.Errorf("unsupported status %q", c.status)
	}
	if err != nil {
		return fmt.Errorf("list requests: %w", err)
	}

	pageSize := c.pageSize
	if pageSize <= 0 {
		cfg, cfgErr := c.rootCmd.RuntimeConfig()
		if cfgErr != nil {
			return fmt.Errorf("runtime config: %w", cfgErr)
		}
		pageSize = cfg.PageSizeDefault
	}
	if cfg, cfgErr := c.rootCmd.RuntimeConfig(); cfgErr == nil {
		if pageSize < cfg.PageSizeMin {
			pageSize = cfg.PageSizeMin
		}
		if pageSize > cfg.PageSizeMax {
			pageSize = cfg.PageSizeMax
		}
	}
	if pageSize < 1 {
		pageSize = 1
	}
	if c.offset < 0 {
		c.offset = 0
	}

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
