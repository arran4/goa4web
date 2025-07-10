package admin

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/utils/netutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/templates"
)

func AdminIPBanPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Bans []*db.BannedIp
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*CoreData)}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.ListBannedIps(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list banned ips: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Bans = rows
	if err := templates.RenderTemplate(w, "iPBanPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminIPBanAddActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	ipNet := strings.TrimSpace(r.PostFormValue("ip"))
	ipNet = netutil.NormalizeIPNet(ipNet)
	reason := strings.TrimSpace(r.PostFormValue("reason"))
	expiresStr := strings.TrimSpace(r.PostFormValue("expires"))
	var expires sql.NullTime
	if expiresStr != "" {
		if t, err := time.Parse("2006-01-02", expiresStr); err == nil {
			expires = sql.NullTime{Time: t, Valid: true}
		}
	}
	if ipNet != "" {
		_ = queries.InsertBannedIp(r.Context(), db.InsertBannedIpParams{
			IpNet:     ipNet,
			Reason:    sql.NullString{String: reason, Valid: reason != ""},
			ExpiresAt: expires,
		})
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminIPBanDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, ip := range r.Form["ip"] {
		ipNet := netutil.NormalizeIPNet(ip)
		if err := queries.CancelBannedIp(r.Context(), ipNet); err != nil {
			log.Printf("cancel banned ip %s: %v", ipNet, err)
		}
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
