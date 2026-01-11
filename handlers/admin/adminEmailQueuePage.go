package admin

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminEmailQueuePage(w http.ResponseWriter, r *http.Request) {
	type EmailItem struct {
		*db.AdminListUnsentPendingEmailsRow
		Email   string
		Subject string
	}
	type Data struct {
		Emails []EmailItem
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Email Queue"
	data := Data{}
	queries := cd.Queries()
	langID, _ := strconv.Atoi(r.URL.Query().Get("lang"))
	role := r.URL.Query().Get("role")
	rows, err := queries.AdminListUnsentPendingEmails(r.Context(), db.AdminListUnsentPendingEmailsParams{
		LanguageID: sql.NullInt32{Int32: int32(langID), Valid: langID != 0},
		RoleName:   role,
	})
	if err != nil {
		log.Printf("list pending emails: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	ids := make([]int32, 0, len(rows))
	for _, e := range rows {
		if e.ToUserID.Valid {
			ids = append(ids, e.ToUserID.Int32)
		}
	}
	users := make(map[int32]*db.SystemGetUserByIDRow)
	for _, id := range ids {
		if u, err := queries.SystemGetUserByID(r.Context(), id); err == nil {
			users[id] = u
		}
	}
	for _, e := range rows {
		emailStr := ""
		if e.ToUserID.Valid && !e.DirectEmail {
			if u, ok := users[e.ToUserID.Int32]; ok && u.Email.Valid && u.Email.String != "" {
				emailStr = u.Email.String
			}
		}
		subj := ""
		if m, err := mail.ReadMessage(strings.NewReader(e.Body)); err == nil {
			if emailStr == "" {
				emailStr = m.Header.Get("To")
			}
			subj = m.Header.Get("Subject")
		}
		if emailStr == "" {
			emailStr = "(unknown)"
		}
		if e.DirectEmail {
			emailStr += " (direct)"
		} else if !e.ToUserID.Valid {
			emailStr += " (userless)"
		}
		data.Emails = append(data.Emails, EmailItem{e, emailStr, subj})
	}
	AdminEmailQueuePageTmpl.Handle(w, r, data)
}

const AdminEmailQueuePageTmpl handlers.Page = "admin/emailQueuePage.gohtml"
