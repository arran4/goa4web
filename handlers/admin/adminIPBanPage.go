package admin

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strings"
	"time"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

// AddIPBanTask blocks a network from accessing the site.
type AddIPBanTask struct{ tasks.TaskString }

// DeleteIPBanTask lifts an IP ban.
type DeleteIPBanTask struct{ tasks.TaskString }

func AdminIPBanPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Bans []*db.BannedIp
	}
	data := Data{CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	rows, err := queries.ListBannedIps(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list banned ips: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Bans = rows
	handlers.TemplateHandler(w, r, "iPBanPage.gohtml", data)
}

func (AddIPBanTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	ipNet := strings.TrimSpace(r.PostFormValue("ip"))
	ipNet = NormalizeIPNet(ipNet)
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
	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (DeleteIPBanTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, ip := range r.Form["ip"] {
		ipNet := NormalizeIPNet(ip)
		if err := queries.CancelBannedIp(r.Context(), ipNet); err != nil {
			log.Printf("cancel banned ip %s: %v", ipNet, err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}
