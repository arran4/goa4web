package admin

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func copyValues(v url.Values) url.Values {
	c := make(url.Values, len(v))
	for k, vals := range v {
		c[k] = append([]string(nil), vals...)
	}
	return c
}

// AdminAuditLogPage shows recent admin actions with basic filtering.
func AdminAuditLogPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Rows     []*db.AdminListAuditLogsRow
		User     string
		Action   string
		PageSize int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Audit Log"
	data := Data{
		User:     r.URL.Query().Get("user"),
		Action:   r.URL.Query().Get("action"),
		PageSize: cd.PageSize(),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	usernameFilter := "%"
	if strings.TrimSpace(data.User) != "" {
		usernameFilter = "%" + data.User + "%"
	}
	actionFilter := "%"
	if strings.TrimSpace(data.Action) != "" {
		actionFilter = "%" + data.Action + "%"
	}

	rows, err := queries.AdminListAuditLogs(r.Context(), db.AdminListAuditLogsParams{
		Username: sql.NullString{String: usernameFilter, Valid: true},
		Action:   actionFilter,
		Limit:    int32(data.PageSize + 1),
		Offset:   int32(offset),
	})
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

	AdminAuditLogPageTmpl.Handle(w, r, data)
}

const AdminAuditLogPageTmpl tasks.Template = "admin/auditLogPage.gohtml"
