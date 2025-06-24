package goa4web

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func copyValues(v url.Values) url.Values {
	c := make(url.Values, len(v))
	for k, vals := range v {
		c[k] = append([]string(nil), vals...)
	}
	return c
}

// adminAuditLogPage shows recent admin actions with basic filtering.
func adminAuditLogPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows     []*ListAuditLogsRow
		User     string
		Action   string
		NextLink string
		PrevLink string
		PageSize int
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		User:     r.URL.Query().Get("user"),
		Action:   r.URL.Query().Get("action"),
		PageSize: getPageSize(r),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	usernameFilter := "%"
	if strings.TrimSpace(data.User) != "" {
		usernameFilter = "%" + data.User + "%"
	}
	actionFilter := "%"
	if strings.TrimSpace(data.Action) != "" {
		actionFilter = "%" + data.Action + "%"
	}

	rows, err := queries.ListAuditLogs(r.Context(), ListAuditLogsParams{
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
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next " + strconv.Itoa(data.PageSize),
			Link: data.NextLink,
		})
	}
	if offset > 0 {
		prevVals := copyValues(params)
		prevVals.Set("offset", strconv.Itoa(offset-data.PageSize))
		data.PrevLink = "/admin/audit?" + prevVals.Encode()
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Previous " + strconv.Itoa(data.PageSize),
			Link: data.PrevLink,
		})
	}

	if err := renderTemplate(w, r, "adminAuditLogPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
