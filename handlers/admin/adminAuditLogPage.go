package admin

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
		*common.CoreData
		Rows     []*db.ListAuditLogsForAdminRow
		User     string
		Action   string
		NextLink string
		PrevLink string
		PageSize int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Audit Log"
	data := Data{
		CoreData: cd,
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

	rows, err := queries.ListAuditLogsForAdmin(r.Context(), db.ListAuditLogsForAdminParams{
		Username: sql.NullString{String: usernameFilter, Valid: true},
		Action:   actionFilter,
		Limit:    int32(data.PageSize + 1),
		Offset:   int32(offset),
	})
	if err != nil {
		log.Printf("list audit logs: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		data.NextLink = "/admin/audit?" + nextVals.Encode()
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Next " + strconv.Itoa(data.PageSize),
			Link: data.NextLink,
		})
	}
	if offset > 0 {
		prevVals := copyValues(params)
		prevVals.Set("offset", strconv.Itoa(offset-data.PageSize))
		data.PrevLink = "/admin/audit?" + prevVals.Encode()
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Previous " + strconv.Itoa(data.PageSize),
			Link: data.PrevLink,
		})
	}

	handlers.TemplateHandler(w, r, "auditLogPage.gohtml", data)
}
