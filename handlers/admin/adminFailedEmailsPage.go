package admin

import (
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminFailedEmailsPage shows queued emails with errors.
func AdminFailedEmailsPage(w http.ResponseWriter, r *http.Request) {
	type EmailItem struct {
		*db.ListFailedEmailsRow
		Email   string
		Subject string
	}
	type Data struct {
		*common.CoreData
		Emails   []EmailItem
		NextLink string
		PrevLink string
		PageSize int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Failed Emails"
	pageSize := cd.PageSize()
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	data := Data{
		CoreData: cd,
		PageSize: pageSize,
	}

	queries := cd.Queries()
	rows, err := queries.ListFailedEmails(r.Context(), db.ListFailedEmailsParams{
		Limit:  int32(pageSize + 1),
		Offset: int32(offset),
	})
	if err != nil {
		log.Printf("list failed emails: %v", err)
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

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
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

	if hasMore {
		data.NextLink = "/admin/email/failed?offset=" + strconv.Itoa(offset+pageSize)
	}
	if offset > 0 {
		prev := offset - pageSize
		if prev < 0 {
			prev = 0
		}
		data.PrevLink = "/admin/email/failed?offset=" + strconv.Itoa(prev)
	}

	handlers.TemplateHandler(w, r, "emailFailedPage.gohtml", data)
}
