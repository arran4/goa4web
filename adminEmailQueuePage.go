package goa4web

import (
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
)

func adminEmailQueuePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Emails []*PendingEmail
	}
	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	items, err := queries.ListUnsentPendingEmails(r.Context())
	if err != nil {
		log.Printf("list pending emails: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Emails = items
	if err := templates.RenderTemplate(w, "emailQueuePage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminEmailQueueResendActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	provider := getEmailProvider()
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		e, err := queries.GetPendingEmailByID(r.Context(), int32(id))
		if err != nil {
			log.Printf("get email: %v", err)
			continue
		}
		if provider != nil {
			if err := provider.Send(r.Context(), e.ToEmail, e.Subject, e.Body); err != nil {
				log.Printf("send email: %v", err)
				continue
			}
		}
		if err := queries.MarkEmailSent(r.Context(), e.ID); err != nil {
			log.Printf("mark sent: %v", err)
		}
	}
	taskDoneAutoRefreshPage(w, r)
}

func adminEmailQueueDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.DeletePendingEmail(r.Context(), int32(id)); err != nil {
			log.Printf("delete email: %v", err)
		}
	}
	taskDoneAutoRefreshPage(w, r)
}
