package goa4web

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/templates"
)

func adminIPBanPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Bans []*BannedIp
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*CoreData)}
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	rows, err := queries.ListBannedIps(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list banned ips: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Bans = rows
	if err := templates.RenderTemplate(w, "iPBanPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminIPBanAddActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	ipNet := strings.TrimSpace(r.PostFormValue("ip"))
	ipNet = normalizeIPNet(ipNet)
	reason := strings.TrimSpace(r.PostFormValue("reason"))
	expiresStr := strings.TrimSpace(r.PostFormValue("expires"))
	var expires sql.NullTime
	if expiresStr != "" {
		if t, err := time.Parse("2006-01-02", expiresStr); err == nil {
			expires = sql.NullTime{Time: t, Valid: true}
		}
	}
	if ipNet != "" {
		_ = queries.InsertBannedIp(r.Context(), InsertBannedIpParams{
			IpNet:     ipNet,
			Reason:    sql.NullString{String: reason, Valid: reason != ""},
			ExpiresAt: expires,
		})
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func adminIPBanDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, ip := range r.Form["ip"] {
		ipNet := normalizeIPNet(ip)
		if err := queries.CancelBannedIp(r.Context(), ipNet); err != nil {
			log.Printf("cancel banned ip %s: %v", ipNet, err)
		}
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
