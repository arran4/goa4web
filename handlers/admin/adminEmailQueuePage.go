package admin

import (
	"github.com/arran4/goa4web/config"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/templates"
)

func AdminEmailQueuePage(w http.ResponseWriter, r *http.Request) {
	type EmailItem struct {
		*db.PendingEmail
		Email   string
		Subject string
	}
	type Data struct {
		*CoreData
		Emails []EmailItem
	}
	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.ListUnsentPendingEmails(r.Context())
	if err != nil {
		log.Printf("list pending emails: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ids := make([]int32, 0, len(rows))
	for _, e := range rows {
		ids = append(ids, e.ToUserID)
	}
	users := make(map[int32]*db.GetUserByIdRow)
	for _, id := range ids {
		if u, err := queries.GetUserById(r.Context(), id); err == nil {
			users[id] = u
		}
	}
	for _, e := range rows {
		emailStr := ""
		if u, ok := users[e.ToUserID]; ok && u.Email.Valid && u.Email.String != "" {
			emailStr = u.Email.String
		}
		subj := ""
		if m, err := mail.ReadMessage(strings.NewReader(e.Body)); err == nil {
			subj = m.Header.Get("Subject")
		}
		data.Emails = append(data.Emails, EmailItem{PendingEmail: e, Email: emailStr, Subject: subj})
	}
	if err := templates.RenderTemplate(w, "emailQueuePage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminEmailQueueResendActionPage(w http.ResponseWriter, r *http.Request) {
	ResendQueueTask.Action()(w, r)
}

func (resendQueueTask) action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	provider := email.ProviderFromConfig(config.AppRuntimeConfig)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	var emails []*db.PendingEmail
	var ids []int32
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		e, err := queries.GetPendingEmailByID(r.Context(), int32(id))
		if err != nil {
			log.Printf("get email: %v", err)
			continue
		}
		emails = append(emails, e)
		ids = append(ids, e.ToUserID)
	}
	users := make(map[int32]*db.GetUserByIdRow)
	for _, id := range ids {
		if u, err := queries.GetUserById(r.Context(), id); err == nil {
			users[id] = u
		}
	}
	for _, e := range emails {
		user, ok := users[e.ToUserID]
		if !ok || !user.Email.Valid || user.Email.String == "" {
			log.Printf("missing or invalid user email for %d", e.ToUserID)
			continue
		}
		addr := mail.Address{Name: user.Username.String, Address: user.Email.String}
		if provider != nil {
			if err := provider.Send(r.Context(), addr, []byte(e.Body)); err != nil {
				log.Printf("send email: %v", err)
				continue
			}
		}
		if err := queries.MarkEmailSent(r.Context(), e.ID); err != nil {
			log.Printf("mark sent: %v", err)
		}
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminEmailQueueDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	DeleteQueueTask.Action()(w, r)
}

func (deleteQueueTask) action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.DeletePendingEmail(r.Context(), int32(id)); err != nil {
			log.Printf("delete email: %v", err)
		}
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
