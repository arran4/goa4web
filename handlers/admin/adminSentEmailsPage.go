package admin

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminSentEmailsPage shows previously sent emails with pagination.
func AdminSentEmailsPage(w http.ResponseWriter, r *http.Request) {
	type EmailItem struct {
		*db.AdminListSentEmailsRow
		Email   string
		Subject string
	}
	type Data struct {
		Emails   []EmailItem
		PageSize int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Sent Emails"
	pageSize := cd.PageSize()
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	data := Data{
		PageSize: pageSize,
	}

	queries := cd.Queries()
	langID, _ := strconv.Atoi(r.URL.Query().Get("lang"))
	role := r.URL.Query().Get("role")
	rows, err := queries.AdminListSentEmails(r.Context(), db.AdminListSentEmailsParams{
		LanguageID: sql.NullInt32{Int32: int32(langID), Valid: langID != 0},
		RoleName:   role,
		Limit:      int32(pageSize + 1),
		Offset:     int32(offset),
	})
	if err != nil {
		log.Printf("list sent emails: %v", err)
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

	params := url.Values{}
	if role != "" {
		params.Set("role", role)
	}
	if langID > 0 {
		params.Set("lang", strconv.Itoa(langID))
	}
	if hasMore {
		nextVals := url.Values{}
		for k, v := range params {
			nextVals[k] = v
		}
		nextVals.Set("offset", strconv.Itoa(offset+pageSize))
		cd.NextLink = "/admin/email/sent?" + nextVals.Encode()
	}
	if offset > 0 {
		prev := offset - pageSize
		if prev < 0 {
			prev = 0
		}
		prevVals := url.Values{}
		for k, v := range params {
			prevVals[k] = v
		}
		prevVals.Set("offset", strconv.Itoa(prev))
		cd.PrevLink = "/admin/email/sent?" + prevVals.Encode()
	}

	AdminEmailSentPageTmpl.Handle(w, r, data)
}

const AdminEmailSentPageTmpl tasks.Template = "admin/emailSentPage.gohtml"
