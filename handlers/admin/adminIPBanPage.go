package admin

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminIPBanPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Bans           []*db.BannedIp
		Error          string
		Notice         string
		ExportURL      string
		ImportTemplate string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "IP Bans"
	data := Data{
		Error:     r.URL.Query().Get("error"),
		Notice:    r.URL.Query().Get("notice"),
		ExportURL: "/admin/ipbans/export",
		ImportTemplate: "action,ip,reason,expires,id\n" +
			"add,203.0.113.0/24,Example ban,2030-01-01,\n" +
			"update,,,2030-12-31,42\n" +
			"delete,198.51.100.11,,,\n",
	}
	queries := cd.Queries()
	rows, err := queries.ListBannedIps(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list banned ips: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	data.Bans = rows
	AdminIPBanPageTmpl.Handle(w, r, data)
}

// AdminIPBanExport streams the current banned IP list as CSV.
func AdminIPBanExport(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	queries := cd.Queries()
	rows, err := queries.ListBannedIps(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list banned ips: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "ip", "reason", "created_at", "expires_at", "canceled_at"})
	for _, row := range rows {
		expires := ""
		if row.ExpiresAt.Valid {
			expires = row.ExpiresAt.Time.Format(time.RFC3339)
		}
		canceled := ""
		if row.CanceledAt.Valid {
			canceled = row.CanceledAt.Time.Format(time.RFC3339)
		}
		_ = writer.Write([]string{
			fmt.Sprintf("%d", row.ID),
			row.IpNet,
			row.Reason.String,
			row.CreatedAt.Format(time.RFC3339),
			expires,
			canceled,
		})
	}
	writer.Flush()
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="ip_bans.csv"`)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("write ip bans csv: %v", err)
	}
}

// AdminIPBanPageTmpl renders the admin IP ban page.
const AdminIPBanPageTmpl tasks.Template = "ipBanPage.gohtml"
