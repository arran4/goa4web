package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminPageTask struct{}

func (t *AdminPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin"
	if _, err := cd.AdminDashboardStats(); err != nil {
		return err
	}
	return AdminPageTmpl.Handler(struct{}{})
}

func (t *AdminPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Admin", "/admin", nil
}

// Ensure interface implementation
var _ tasks.Task = (*AdminPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminPageTask)(nil)

type AdminRolesPageTask struct{}

func (t *AdminRolesPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		Roles []*db.Role
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Roles"
	roles, err := cd.AllRoles()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	data := Data{Roles: roles}
	return AdminRolesPageTmpl.Handler(data)
}

func (t *AdminRolesPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Roles", "/admin/roles", &AdminPageTask{}
}

var _ tasks.Task = (*AdminRolesPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminRolesPageTask)(nil)

type AdminRolePageBreadcrumb struct {
	RoleName string
	RoleID   int32
}

func (p *AdminRolePageBreadcrumb) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Role %s", p.RoleName), "", &AdminRolesPageTask{}
}

type AdminAnnouncementsPageTask struct{}

func (t *AdminAnnouncementsPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		Announcements []*db.AdminListAnnouncementsWithNewsRow
		NewsID        string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Announcements"
	data := Data{}
	queries := cd.Queries()
	rows, err := queries.AdminListAnnouncementsWithNews(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	data.Announcements = rows
	data.NewsID = r.FormValue("news_id")
	return AdminAnnouncementsPageTmpl.Handler(data)
}

func (t *AdminAnnouncementsPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Announcements", "/admin/announcements", &AdminPageTask{}
}

var _ tasks.Task = (*AdminAnnouncementsPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminAnnouncementsPageTask)(nil)

type AdminCommentsPageTask struct{}

func (t *AdminCommentsPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Comments"
	queries := cd.Queries()
	rows, err := queries.AdminListAllCommentsWithThreadInfo(r.Context(), db.AdminListAllCommentsWithThreadInfoParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		return err
	}
	data := struct {
		*common.CoreData
		Comments []*db.AdminListAllCommentsWithThreadInfoRow
	}{cd, rows}
	return AdminCommentsPageTmpl.Handler(data)
}

func (t *AdminCommentsPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Comments", "/admin/comments", &AdminPageTask{}
}

var _ tasks.Task = (*AdminCommentsPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminCommentsPageTask)(nil)

type AdminCommentPageBreadcrumb struct {
	CommentID int32
}

func (p *AdminCommentPageBreadcrumb) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Comment %d", p.CommentID), "", &AdminCommentsPageTask{}
}

type AdminEmailPageTask struct{}

func (t *AdminEmailPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	filters := emailFiltersFromRequest(r)

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
		totalCount, err = queries.AdminCountSentEmails(r.Context(), db.AdminCountSentEmailsParams{
			Provider:      filters.ProviderParam(),
			CreatedBefore: filters.CreatedBefore, // Uses SentAt
			LanguageID:    filters.LangIDParam(),
			RoleName:      filters.RoleParam(),
		})
		if err != nil {
			return err
		}

		var dbRows []*db.AdminListSentEmailsRow
		dbRows, err = queries.AdminListSentEmails(r.Context(), db.AdminListSentEmailsParams{
			Provider:      filters.ProviderParam(),
			CreatedBefore: filters.CreatedBefore,
			LanguageID:    filters.LangIDParam(),
			RoleName:      filters.RoleParam(),
			Limit:         int32(pageSize + 1),
			Offset:        int32(offset),
		})
		if err == nil {
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
		}
	} else {
		// Queue or Failed
		totalCount, err = queries.AdminCountUnsentPendingEmails(r.Context(), db.AdminCountUnsentPendingEmailsParams{
			Status:        filters.StatusParam(),
			Provider:      filters.ProviderParam(),
			CreatedBefore: filters.CreatedBefore,
			LanguageID:    filters.LangIDParam(),
			RoleName:      filters.RoleParam(),
		})
		if err != nil {
			return err
		}

		var dbRows []*db.AdminListUnsentPendingEmailsRow
		dbRows, err = queries.AdminListUnsentPendingEmails(r.Context(), db.AdminListUnsentPendingEmailsParams{
			Status:        filters.StatusParam(),
			Provider:      filters.ProviderParam(),
			CreatedBefore: filters.CreatedBefore,
			LanguageID:    filters.LangIDParam(),
			RoleName:      filters.RoleParam(),
			Limit:         int32(pageSize + 1),
			Offset:        int32(offset),
		})
		if err == nil {
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
		}
	}

	if err != nil {
		return err
	}

	data.FilteredCount = totalCount

	// Pagination logic (if fetched one more than pageSize)
	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}

	// Resolve Users
	ids := make([]int32, 0, len(rows))
	rowIDs := make([]int32, 0, len(rows))
	for _, e := range rows {
		rowIDs = append(rowIDs, e.ID)
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
	data.Emails = rows
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
		for k, v := range params {
			nextVals[k] = v
		}
		nextVals.Set("offset", strconv.Itoa(offset+pageSize))
		cd.NextLink = r.URL.Path + "?" + nextVals.Encode()
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
		cd.PrevLink = r.URL.Path + "?" + prevVals.Encode()
	}

	return AdminEmailPageTmpl.Handler(data)
}

func (t *AdminEmailPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Email", "/admin/email/queue", &AdminPageTask{}
}

var _ tasks.Task = (*AdminEmailPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminEmailPageTask)(nil)

type AdminUserListPageTask struct{}

func (t *AdminUserListPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Users"
	if _, err := cd.AdminListUsers(); err != nil {
		return err
	}
	return AdminUserListPageTmpl.Handler(struct{}{})
}

func (t *AdminUserListPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Users", "/admin/user", &AdminPageTask{}
}

var _ tasks.Task = (*AdminUserListPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminUserListPageTask)(nil)

type AdminUserProfilePageBreadcrumb struct {
	UserID   int32
	UserName string
}

func (p *AdminUserProfilePageBreadcrumb) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	title := fmt.Sprintf("User %d", p.UserID)
	if p.UserName != "" {
		title = fmt.Sprintf("User %s", p.UserName)
	}
	return title, "", &AdminUserListPageTask{}
}

type AdminSiteSettingsPageTask struct {
	ConfigFile string
}

func (t *AdminSiteSettingsPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Site Settings"
	cd.FeedsEnabled = cd.Config.FeedsEnabled

	values := config.ValuesMap(*cd.Config)
	defaults := config.DefaultMap(config.NewRuntimeConfig())
	usages := config.UsageMap()
	examples := config.ExamplesMap()
	flags := config.NameMap()

	fileVals, _ := config.LoadAppConfigFile(core.OSFS{}, t.ConfigFile)
	hide := map[string]struct{}{
		config.EnvDBConn:              {},
		config.EnvSMTPPass:            {},
		config.EnvJMAPPass:            {},
		config.EnvSendGridKey:         {},
		config.EnvSessionSecret:       {},
		config.EnvSessionSecretFile:   {},
		config.EnvImageSignSecret:     {},
		config.EnvImageSignSecretFile: {},
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		if _, ok := hide[k]; ok {
			delete(values, k)
			continue
		}
		if values[k] == "" {
			delete(values, k)
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	type detail struct {
		Env     string
		Flag    string
		Value   string
		Default string
		Usage   string
		Example []string
		Source  string
	}
	cfg := make([]detail, 0, len(keys))
	for _, k := range keys {
		src := "default"
		if v := fileVals[k]; v != "" && v == values[k] {
			src = "file"
		} else if v := os.Getenv(k); v != "" && v == values[k] {
			src = "env"
		} else if values[k] != defaults[k] {
			src = "flag"
		}
		cfg = append(cfg, detail{
			Env:     k,
			Flag:    flags[k],
			Value:   values[k],
			Default: defaults[k],
			Usage:   usages[k],
			Example: examples[k],
			Source:  src,
		})
	}

	data := struct {
		ConfigFile string
		Config     []detail
	}{
		ConfigFile: t.ConfigFile,
		Config:     cfg,
	}

	return AdminSiteSettingsPageTmpl.Handler(data)
}

func (t *AdminSiteSettingsPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Site Settings", "/admin/settings", &AdminPageTask{}
}

var _ tasks.Task = (*AdminSiteSettingsPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminSiteSettingsPageTask)(nil)

type AdminExternalLinksPageTask struct{}

func (t *AdminExternalLinksPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		Links         []*db.ExternalLink
		Query         string
		ResultSummary string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "External Links"
	query := strings.TrimSpace(r.URL.Query().Get(externalLinksFilterQueryParam))
	queries := cd.Queries()
	rows, err := queries.AdminListExternalLinks(r.Context(), db.AdminListExternalLinksParams{Limit: 200, Offset: 0})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if query != "" {
		filtered := make([]*db.ExternalLink, 0, len(rows))
		queryLower := strings.ToLower(query)
		for _, link := range rows {
			if externalLinkMatchesQuery(link, queryLower) {
				filtered = append(filtered, link)
			}
		}
		rows = filtered
	}
	data := Data{
		Links: rows,
		Query: query,
	}
	action := r.URL.Query().Get(externalLinksActionQueryParam)
	successCount := queryIntValue(r, externalLinksSuccessQueryParam)
	failureCount := queryIntValue(r, externalLinksFailureQueryParam)
	if action != "" {
		actionLabel := action
		switch action {
		case externalLinksActionRefresh:
			actionLabel = "Refreshed"
		case externalLinksActionDelete:
			actionLabel = "Deleted"
		}
		data.ResultSummary = fmt.Sprintf("%s external links: %d succeeded, %d failed.", actionLabel, successCount, failureCount)
	}
	return AdminExternalLinksPageTmpl.Handler(data)
}

func (t *AdminExternalLinksPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "External Links", "/admin/external-links", &AdminPageTask{}
}

var _ tasks.Task = (*AdminExternalLinksPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminExternalLinksPageTask)(nil)

type AdminExternalLinkDetailsPageBreadcrumb struct {
	LinkID int32
}

func (p *AdminExternalLinkDetailsPageBreadcrumb) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Link %d", p.LinkID), "", &AdminExternalLinksPageTask{}
}
