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

// AdminFailedEmailsPage shows queued emails with errors.
func AdminFailedEmailsPage(w http.ResponseWriter, r *http.Request) {
	type EmailItem struct {
		*db.AdminListFailedEmailsRow
		Email   string
		Subject string
	}
	type Data struct {
		Emails     []EmailItem
		PageSize   int
		StatusByID map[int32]string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Failed Emails"
	pageSize := cd.PageSize()
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	data := Data{
		PageSize: pageSize,
	}

	queries := cd.Queries()
	langID, _ := strconv.Atoi(r.URL.Query().Get("lang"))
	role := r.URL.Query().Get("role")
	rows, err := queries.AdminListFailedEmails(r.Context(), db.AdminListFailedEmailsParams{
		LanguageID: sql.NullInt32{Int32: int32(langID), Valid: langID != 0},
		RoleName:   role,
		Limit:      int32(pageSize + 1),
		Offset:     int32(offset),
	})
	if err != nil {
		log.Printf("list failed emails: %v", err)
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

	rowIDs := make([]int32, 0, len(rows))
	for _, e := range rows {
		rowIDs = append(rowIDs, e.ID)
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
	data.StatusByID = buildEmailStatusMap(r, rowIDs)

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
		cd.NextLink = "/admin/email/failed?" + nextVals.Encode()
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
		cd.PrevLink = "/admin/email/failed?" + prevVals.Encode()
	}

	AdminEmailFailedPageTmpl.Handle(w, r, data)
}

const AdminEmailFailedPageTmpl tasks.Template = "admin/emailFailedPage.gohtml"
