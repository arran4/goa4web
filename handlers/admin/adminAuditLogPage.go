package admin

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
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

func copyValues(v url.Values) url.Values {
	c := make(url.Values, len(v))
	for k, vals := range v {
		c[k] = append([]string(nil), vals...)
	}
	return c
}

const (
	// auditLogDateLayout defines the date-only format accepted by the audit log filters.
	auditLogDateLayout = "2006-01-02"
	// auditLogDateTimeLayout defines the datetime format accepted by the audit log filters.
	auditLogDateTimeLayout = "2006-01-02T15:04"
)

func parseAuditLogTime(value string, isEnd bool) (sql.NullTime, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullTime{}, ""
	}
	if parsed, err := time.ParseInLocation(auditLogDateTimeLayout, value, time.Local); err == nil {
		return sql.NullTime{Time: parsed, Valid: true}, value
	}
	if parsed, err := time.ParseInLocation(auditLogDateLayout, value, time.Local); err == nil {
		if isEnd {
			parsed = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		return sql.NullTime{Time: parsed, Valid: true}, parsed.Format(auditLogDateTimeLayout)
	}
	return sql.NullTime{}, value
}

// AdminAuditLogPage shows recent admin actions with basic filtering.
func AdminAuditLogPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Rows       []*db.AdminListAuditLogsRow
		User       string
		Action     string
		Section    string
		StartTime  string
		EndTime    string
		PageSize   int
		Summary    []*db.AdminAuditLogActionSummaryRow
		ExportLink string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Audit Log"
	query := r.URL.Query()
	data := Data{
		User:     query.Get("user"),
		Action:   query.Get("action"),
		Section:  query.Get("section"),
		PageSize: cd.PageSize(),
	}

	startTime, startValue := parseAuditLogTime(query.Get("start"), false)
	endTime, endValue := parseAuditLogTime(query.Get("end"), true)
	data.StartTime = startValue
	data.EndTime = endValue

	offset, _ := strconv.Atoi(query.Get("offset"))
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	usernameFilter := "%"
	if strings.TrimSpace(data.User) != "" {
		usernameFilter = "%" + data.User + "%"
	}
	actionFilter := "%"
	if strings.TrimSpace(data.Action) != "" {
		actionFilter = "%" + data.Action + "%"
	}
	sectionFilter := "%"
	if strings.TrimSpace(data.Section) != "" {
		sectionFilter = "%" + data.Section + "%"
	}

	listParams := db.AdminListAuditLogsParams{
		Username:  sql.NullString{String: usernameFilter, Valid: true},
		Action:    actionFilter,
		Section:   sectionFilter,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     int32(data.PageSize + 1),
		Offset:    int32(offset),
	}

	if query.Get("export") == "csv" {
		rows, err := queries.AdminListAuditLogs(r.Context(), listParams)
		if err != nil {
			log.Printf("list audit logs for export: %v", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
		writeAuditLogCSV(w, rows)
		return
	}

	rows, err := queries.AdminListAuditLogs(r.Context(), listParams)
	if err != nil {
		log.Printf("list audit logs: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	hasMore := len(rows) > data.PageSize
	if hasMore {
		rows = rows[:data.PageSize]
	}
	data.Rows = rows

	params := url.Values{}
	if data.User != "" {
		params.Set("user", data.User)
	}
	if data.Action != "" {
		params.Set("action", data.Action)
	}
	if data.Section != "" {
		params.Set("section", data.Section)
	}
	if data.StartTime != "" {
		params.Set("start", data.StartTime)
	}
	if data.EndTime != "" {
		params.Set("end", data.EndTime)
	}
	encodedParams := params.Encode()
	if encodedParams == "" {
		data.ExportLink = "/admin/audit?export=csv"
	} else {
		data.ExportLink = "/admin/audit?" + encodedParams + "&export=csv"
	}

	if hasMore {
		nextVals := copyValues(params)
		nextVals.Set("offset", strconv.Itoa(offset+data.PageSize))
		cd.NextLink = "/admin/audit?" + nextVals.Encode()
	}
	if offset > 0 {
		prevVals := copyValues(params)
		prevVals.Set("offset", strconv.Itoa(offset-data.PageSize))
		cd.PrevLink = "/admin/audit?" + prevVals.Encode()
	}

	summaryRows, err := queries.AdminAuditLogActionSummary(r.Context(), db.AdminAuditLogActionSummaryParams{
		Username:  sql.NullString{String: usernameFilter, Valid: true},
		Action:    actionFilter,
		Section:   sectionFilter,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		log.Printf("list audit log summary: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Summary = summaryRows

	AdminAuditLogPageTmpl.Handle(w, r, data)
}

const AdminAuditLogPageTmpl tasks.Template = "admin/auditLogPage.gohtml"

func writeAuditLogCSV(w http.ResponseWriter, rows []*db.AdminListAuditLogsRow) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"audit-log.csv\"")
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"ID", "User", "Action", "Path", "Details", "CreatedAt"})
	for _, row := range rows {
		username := ""
		if row.Username.Valid {
			username = row.Username.String
		}
		details := ""
		if row.Details.Valid {
			details = row.Details.String
		}
		_ = writer.Write([]string{
			strconv.Itoa(int(row.ID)),
			username,
			row.Action,
			row.Path,
			details,
			row.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()
}
