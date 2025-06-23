package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
)

func adminIPBanPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Bans []*BannedIp
	}
	data := Data{CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData)}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	rows, err := queries.ListBannedIps(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list banned ips: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Bans = rows
	if err := renderTemplate(w, r, "adminIPBanPage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminIPBanAddActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	ip := strings.TrimSpace(r.PostFormValue("ip"))
	reason := strings.TrimSpace(r.PostFormValue("reason"))
	if ip != "" {
		_ = queries.InsertBannedIp(r.Context(), InsertBannedIpParams{
			IpAddress: ip,
			Reason:    sql.NullString{String: reason, Valid: reason != ""},
		})
	}
	taskDoneAutoRefreshPage(w, r)
}

func adminIPBanDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, ip := range r.Form["ip"] {
		if err := queries.DeleteBannedIp(r.Context(), ip); err != nil {
			log.Printf("delete banned ip %s: %v", ip, err)
		}
	}
	taskDoneAutoRefreshPage(w, r)
}
