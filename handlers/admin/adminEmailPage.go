package admin

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"maps"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type EmailItem struct {
	ID          int32
	ToUserID    sql.NullInt32
	Body        string
	ErrorCount  int32
	CreatedAt   time.Time
	SentAt      sql.NullTime
	DirectEmail bool
	EmailStr    string
	Subject     string
}

// AdminEmailPage handles queue, failed, and sent email views.
func AdminEmailPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	filters := emailFiltersFromValues(r.URL.Query())

	mode := "queue"
	if strings.Contains(r.URL.Path, "/failed") {
		mode = "failed"
		filters.Status = "failed" // Force status=failed for this view
	} else if strings.Contains(r.URL.Path, "/sent") {
		mode = "sent"
	}

	cd.PageTitle = fmt.Sprintf("Email %s", strings.Title(mode))

	pageSize := cd.PageSize()
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	type Data struct {
		Emails        []EmailItem
		Filters       EmailFilters
		FilteredCount int64
		Mode          string
		PageSize      int
		StatusByID    map[int32]string
	}

	data := Data{
		Filters:  filters,
		Mode:     mode,
		PageSize: pageSize,
	}

	var rows []EmailItem
	var totalCount int64
	var err error

	if mode == "sent" {
		rows, totalCount, err = fetchSentEmails(r.Context(), queries, filters, pageSize, offset)
	} else {
		rows, totalCount, err = fetchUnsentPendingEmails(r.Context(), queries, filters, pageSize, offset)
	}

	if err != nil {
		log.Printf("list emails: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	data.FilteredCount = totalCount

	// Pagination logic (if fetched one more than pageSize)
	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}

	rows = decorateEmailItems(r.Context(), queries, rows)
	data.Emails = rows

	rowIDs := make([]int32, 0, len(rows))
	for _, e := range rows {
		rowIDs = append(rowIDs, e.ID)
	}
	data.StatusByID = buildEmailStatusMap(r, rowIDs)

	// Pagination Links
	params := url.Values{}
	// Copy params from request but exclude offset
	for k, v := range r.URL.Query() {
		if k != "offset" {
			params[k] = v
		}
	}

	if hasMore {
		nextVals := url.Values{}
		maps.Copy(nextVals, params)
		nextVals.Set("offset", strconv.Itoa(offset+pageSize))
		cd.NextLink = r.URL.Path + "?" + nextVals.Encode()
	}
	if offset > 0 {
		prev := max(offset-pageSize, 0)
		prevVals := url.Values{}
		maps.Copy(prevVals, params)
		prevVals.Set("offset", strconv.Itoa(prev))
		cd.PrevLink = r.URL.Path + "?" + prevVals.Encode()
	}

	AdminEmailPageTmpl.Handle(w, r, data)
}

func fetchSentEmails(ctx context.Context, queries db.Querier, filters EmailFilters, pageSize, offset int) ([]EmailItem, int64, error) {
	totalCount, err := queries.AdminCountSentEmails(ctx, db.AdminCountSentEmailsParams{
		Provider:      filters.ProviderParam(),
		CreatedBefore: filters.CreatedBefore, // Uses SentAt
		LanguageID:    filters.LangIDParam(),
		RoleName:      filters.RoleParam(),
	})
	if err != nil {
		return nil, 0, err
	}

	dbRows, err := queries.AdminListSentEmails(ctx, db.AdminListSentEmailsParams{
		Provider:      filters.ProviderParam(),
		CreatedBefore: filters.CreatedBefore,
		LanguageID:    filters.LangIDParam(),
		RoleName:      filters.RoleParam(),
		Limit:         int32(pageSize + 1),
		Offset:        int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	var rows []EmailItem
	for _, r := range dbRows {
		rows = append(rows, EmailItem{
			ID:          r.ID,
			ToUserID:    r.ToUserID,
			Body:        r.Body,
			ErrorCount:  r.ErrorCount,
			CreatedAt:   r.CreatedAt,
			SentAt:      r.SentAt,
			DirectEmail: r.DirectEmail,
		})
	}
	return rows, totalCount, nil
}

func fetchUnsentPendingEmails(ctx context.Context, queries db.Querier, filters EmailFilters, pageSize, offset int) ([]EmailItem, int64, error) {
	totalCount, err := queries.AdminCountUnsentPendingEmails(ctx, db.AdminCountUnsentPendingEmailsParams{
		Status:        filters.StatusParam(),
		Provider:      filters.ProviderParam(),
		CreatedBefore: filters.CreatedBefore,
		LanguageID:    filters.LangIDParam(),
		RoleName:      filters.RoleParam(),
	})
	if err != nil {
		return nil, 0, err
	}

	dbRows, err := queries.AdminListUnsentPendingEmails(ctx, db.AdminListUnsentPendingEmailsParams{
		Status:        filters.StatusParam(),
		Provider:      filters.ProviderParam(),
		CreatedBefore: filters.CreatedBefore,
		LanguageID:    filters.LangIDParam(),
		RoleName:      filters.RoleParam(),
		Limit:         int32(pageSize + 1),
		Offset:        int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	var rows []EmailItem
	for _, r := range dbRows {
		rows = append(rows, EmailItem{
			ID:          r.ID,
			ToUserID:    r.ToUserID,
			Body:        r.Body,
			ErrorCount:  r.ErrorCount,
			CreatedAt:   r.CreatedAt,
			DirectEmail: r.DirectEmail,
		})
	}
	return rows, totalCount, nil
}

func decorateEmailItems(ctx context.Context, queries db.Querier, rows []EmailItem) []EmailItem {
	// Resolve Users
	ids := make([]int32, 0, len(rows))
	for _, e := range rows {
		if e.ToUserID.Valid {
			ids = append(ids, e.ToUserID.Int32)
		}
	}
	users := make(map[int32]*db.SystemGetUsersByIDsRow)
	if len(ids) > 0 {
		if us, err := queries.SystemGetUsersByIDs(ctx, ids); err == nil {
			for _, u := range us {
				users[u.Idusers] = u
			}
		}
	}

	// Decorate items
	for i, e := range rows {
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
		rows[i].EmailStr = emailStr
		rows[i].Subject = subj
	}
	return rows
}

const AdminEmailPageTmpl tasks.Template = "admin/emailPage.gohtml"
