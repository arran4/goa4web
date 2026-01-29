package admin

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// ipBanBulkDateLayout defines the date format for CSV expires values.
const ipBanBulkDateLayout = "2006-01-02"

// IPBanBulkTask applies bulk IP ban actions from CSV.
type IPBanBulkTask struct{ tasks.TaskString }

var ipBanBulkTask = &IPBanBulkTask{TaskString: TaskBulkImport}

var _ tasks.Task = (*IPBanBulkTask)(nil)
var _ tasks.AuditableTask = (*IPBanBulkTask)(nil)

type ipBanBulkRow struct {
	action  string
	ipNet   string
	reason  string
	expires sql.NullTime
	id      int32
	line    int
}

func (IPBanBulkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	csvText := strings.TrimSpace(r.PostFormValue("csv"))
	if csvText == "" {
		return fmt.Errorf("csv required %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("csv required")))
	}
	rows, validationErrors, err := parseIPBanBulkCSV(csvText)
	if err != nil {
		return fmt.Errorf("read csv fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if len(validationErrors) > 0 {
		return fmt.Errorf("validate csv fail %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("%s", summarizeIPBanBulkErrors(validationErrors))))
	}
	adds, updates, deletes := countIPBanBulkActions(rows)
	if r.PostFormValue("dry_run") != "" {
		notice := fmt.Sprintf("Dry run complete: %d add, %d update, %d delete", adds, updates, deletes)
		trackIPBanBulkEvent(cd, adds, updates, deletes, true)
		return handlers.RefreshDirectHandler{TargetURL: ipBanBulkNoticeURL(notice)}
	}
	queries := cd.Queries()
	for _, row := range rows {
		switch row.action {
		case "add":
			if err := queries.AdminInsertBannedIp(r.Context(), db.AdminInsertBannedIpParams{
				IpNet:     row.ipNet,
				Reason:    sql.NullString{String: row.reason, Valid: row.reason != ""},
				ExpiresAt: row.expires,
			}); err != nil {
				return fmt.Errorf("insert banned ip %s fail %w", row.ipNet, handlers.ErrRedirectOnSamePageHandler(err))
			}
		case "delete":
			if err := queries.AdminCancelBannedIp(r.Context(), row.ipNet); err != nil {
				return fmt.Errorf("cancel banned ip %s fail %w", row.ipNet, handlers.ErrRedirectOnSamePageHandler(err))
			}
		case "update":
			if err := queries.AdminUpdateBannedIp(r.Context(), db.AdminUpdateBannedIpParams{
				Reason:    sql.NullString{String: row.reason, Valid: row.reason != ""},
				ExpiresAt: row.expires,
				ID:        row.id,
			}); err != nil {
				return fmt.Errorf("update banned ip %d fail %w", row.id, handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
	}
	trackIPBanBulkEvent(cd, adds, updates, deletes, false)
	notice := fmt.Sprintf("Applied bulk import: %d add, %d update, %d delete", adds, updates, deletes)
	return handlers.RefreshDirectHandler{TargetURL: ipBanBulkNoticeURL(notice)}
}

// AuditRecord summarises a bulk IP ban import.
func (IPBanBulkTask) AuditRecord(data map[string]any) string {
	adds := readIPBanBulkCount(data["Adds"])
	updates := readIPBanBulkCount(data["Updates"])
	deletes := readIPBanBulkCount(data["Deletes"])
	mod, _ := data["Moderator"].(string)
	if mod == "" {
		mod = "admin"
	}
	if dry, _ := data["DryRun"].(bool); dry {
		return fmt.Sprintf("%s ran a dry run for IP bans (%d add, %d update, %d delete)", mod, adds, updates, deletes)
	}
	return fmt.Sprintf("%s bulk updated IP bans (%d add, %d update, %d delete)", mod, adds, updates, deletes)
}

func parseIPBanBulkCSV(csvText string) ([]ipBanBulkRow, []string, error) {
	reader := csv.NewReader(strings.NewReader(csvText))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) == 0 {
		return nil, []string{"no CSV rows found"}, nil
	}
	header, startIndex := resolveIPBanBulkHeader(records)
	var rows []ipBanBulkRow
	var errs []string
	for i := startIndex; i < len(records); i++ {
		rec := records[i]
		if isIPBanBulkRowEmpty(rec) {
			continue
		}
		action := strings.ToLower(strings.TrimSpace(ipBanBulkColumnValue(rec, header, "action", 0)))
		ipNet := strings.TrimSpace(ipBanBulkColumnValue(rec, header, "ip", 1))
		reason := strings.TrimSpace(ipBanBulkColumnValue(rec, header, "reason", 2))
		expiresStr := strings.TrimSpace(ipBanBulkColumnValue(rec, header, "expires", 3))
		idStr := strings.TrimSpace(ipBanBulkColumnValue(rec, header, "id", 4))
		line := i + 1
		if action == "" {
			errs = append(errs, fmt.Sprintf("row %d: action required", line))
			continue
		}
		if action != "add" && action != "delete" && action != "update" {
			errs = append(errs, fmt.Sprintf("row %d: invalid action %q", line, action))
			continue
		}
		var expires sql.NullTime
		if expiresStr != "" {
			t, err := time.Parse(ipBanBulkDateLayout, expiresStr)
			if err != nil {
				errs = append(errs, fmt.Sprintf("row %d: invalid expires date %q", line, expiresStr))
				continue
			}
			expires = sql.NullTime{Time: t, Valid: true}
		}
		var id int32
		if idStr != "" {
			val, err := strconv.Atoi(idStr)
			if err != nil || val <= 0 {
				errs = append(errs, fmt.Sprintf("row %d: invalid id %q", line, idStr))
				continue
			}
			id = int32(val)
		}
		switch action {
		case "add", "delete":
			ipNet = NormalizeIPNet(ipNet)
			if ipNet == "" {
				errs = append(errs, fmt.Sprintf("row %d: ip required", line))
				continue
			}
		case "update":
			if id == 0 {
				errs = append(errs, fmt.Sprintf("row %d: id required for update", line))
				continue
			}
		}
		rows = append(rows, ipBanBulkRow{
			action:  action,
			ipNet:   ipNet,
			reason:  reason,
			expires: expires,
			id:      id,
			line:    line,
		})
	}
	if len(rows) == 0 && len(errs) == 0 {
		errs = append(errs, "no actionable rows found")
	}
	return rows, errs, nil
}

func resolveIPBanBulkHeader(records [][]string) (map[string]int, int) {
	if len(records) == 0 {
		return nil, 0
	}
	headerRow := records[0]
	header := map[string]int{}
	for idx, raw := range headerRow {
		key := strings.ToLower(strings.TrimSpace(raw))
		switch key {
		case "action":
			header["action"] = idx
		case "ip", "ip_net", "cidr":
			header["ip"] = idx
		case "reason", "ban_reason":
			header["reason"] = idx
		case "expires", "expires_at":
			header["expires"] = idx
		case "id", "ban_id":
			header["id"] = idx
		}
	}
	if len(header) == 0 {
		return nil, 0
	}
	return header, 1
}

func ipBanBulkColumnValue(row []string, header map[string]int, name string, fallback int) string {
	index := fallback
	if header != nil {
		if value, ok := header[name]; ok {
			index = value
		} else {
			return ""
		}
	}
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

func isIPBanBulkRowEmpty(row []string) bool {
	for _, val := range row {
		if strings.TrimSpace(val) != "" {
			return false
		}
	}
	return true
}

func summarizeIPBanBulkErrors(errs []string) string {
	if len(errs) <= 5 {
		return strings.Join(errs, "; ")
	}
	return fmt.Sprintf("%s; and %d more", strings.Join(errs[:5], "; "), len(errs)-5)
}

func countIPBanBulkActions(rows []ipBanBulkRow) (adds, updates, deletes int) {
	for _, row := range rows {
		switch row.action {
		case "add":
			adds++
		case "update":
			updates++
		case "delete":
			deletes++
		}
	}
	return adds, updates, deletes
}

func ipBanBulkNoticeURL(notice string) string {
	return "/admin/ipbans?notice=" + url.QueryEscape(notice)
}

func trackIPBanBulkEvent(cd *common.CoreData, adds, updates, deletes int, dryRun bool) {
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["Adds"] = adds
		evt.Data["Updates"] = updates
		evt.Data["Deletes"] = deletes
		evt.Data["DryRun"] = dryRun
		if u, _ := cd.CurrentUser(); u != nil && u.Username.Valid {
			evt.Data["Moderator"] = u.Username.String
		}
	}
}

func readIPBanBulkCount(val any) int {
	switch v := val.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}
