package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminEmailQueuePage(w http.ResponseWriter, r *http.Request) {
	type EmailItem struct {
		*db.ListUnsentPendingEmailsRow
		Email   string
		Subject string
	}
	type Data struct {
		*common.CoreData
		Emails []EmailItem
	}
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	rows, err := queries.ListUnsentPendingEmails(r.Context())
	if err != nil {
		log.Printf("list pending emails: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ids := make([]int32, 0, len(rows))
	for _, e := range rows {
		if e.ToUserID.Valid {
			ids = append(ids, e.ToUserID.Int32)
		}
	}
	users := make(map[int32]*db.GetUserByIdRow)
	for _, id := range ids {
		if u, err := queries.GetUserById(r.Context(), id); err == nil {
			users[id] = u
		}
	}
	for _, e := range rows {
		emailStr := ""
		if e.ToUserID.Valid {
			if u, ok := users[e.ToUserID.Int32]; ok && u.Email.Valid && u.Email.String != "" {
				emailStr = u.Email.String
			}
		}
		subj := ""
		if m, err := mail.ReadMessage(strings.NewReader(e.Body)); err == nil {
			subj = m.Header.Get("Subject")
		}
		data.Emails = append(data.Emails, EmailItem{e, emailStr, subj})
	}
	handlers.TemplateHandler(w, r, "emailQueuePage.gohtml", data)
}
