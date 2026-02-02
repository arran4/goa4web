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
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type AdminPage struct{}

func (p *AdminPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin"
	if _, err := cd.AdminDashboardStats(); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	AdminPageTmpl.Handler(struct{}{}).ServeHTTP(w, r)
}

func (p *AdminPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Admin", "/admin", nil
}

func (p *AdminPage) PageTitle() string {
	return "Admin"
}

// Ensure interface implementation
var _ tasks.Page = (*AdminPage)(nil)
var _ tasks.Task = (*AdminPage)(nil)
var _ http.Handler = (*AdminPage)(nil)

type AdminRolesPage struct{}

func (p *AdminRolesPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminRolesPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Roles []*db.Role
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Roles"
	roles, err := cd.AllRoles()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	data := Data{Roles: roles}
	AdminRolesPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminRolesPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Roles", "/admin/roles", &AdminPage{}
}

func (p *AdminRolesPage) PageTitle() string {
	return "Admin Roles"
}

var _ tasks.Page = (*AdminRolesPage)(nil)
var _ tasks.Task = (*AdminRolesPage)(nil)
var _ http.Handler = (*AdminRolesPage)(nil)

type AdminRolePage struct {
	RoleName string
	RoleID   int32
	Data     any
}

func (p *AdminRolePage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Role %s", p.RoleName), "", &AdminRolesPage{}
}

func (p *AdminRolePage) PageTitle() string {
	return fmt.Sprintf("Role: %s", p.RoleName)
}

func (p *AdminRolePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AdminRolePageTmpl.Handler(p.Data).ServeHTTP(w, r)
}

var _ tasks.Page = (*AdminRolePage)(nil)
var _ http.Handler = (*AdminRolePage)(nil)

type AdminRoleTask struct{}

func (t *AdminRoleTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	role, err := cd.SelectedRole()
	if err != nil || role == nil {
		return fmt.Errorf("role not found")
	}

	id := cd.SelectedRoleID()
	emailRows, err := queries.GetVerifiedUserEmails(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	emailsByUser := make(map[int32][]string)
	for _, row := range emailRows {
		emailsByUser[row.UserID] = append(emailsByUser[row.UserID], row.Email)
	}

	users, err := queries.AdminListUsersByRoleID(r.Context(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	roleUsers := make([]*roleUser, 0, len(users))
	for _, u := range users {
		ru := &roleUser{ID: u.Idusers, User: u.Username, UserID: u.Idusers}
		if emails, ok := emailsByUser[u.Idusers]; ok {
			ru.Email = emails
		}
		roleUsers = append(roleUsers, ru)
	}

	groups, err := buildGrantGroups(r.Context(), cd, id)
	if err != nil {
		return err
	}

	data := struct {
		Role        *db.Role
		Users       []*roleUser
		GrantGroups []GrantGroup
	}{
		Role:        role,
		Users:       roleUsers,
		GrantGroups: groups,
	}

	return &AdminRolePage{
		RoleName: role.Name,
		RoleID:   role.ID,
		Data:     data,
	}
}

type AdminAnnouncementsPage struct{}

func (p *AdminAnnouncementsPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminAnnouncementsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		handlers.RenderErrorPage(w, r, err)
		return
	}
	data.Announcements = rows
	data.NewsID = r.FormValue("news_id")
	AdminAnnouncementsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminAnnouncementsPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Announcements", "/admin/announcements", &AdminPage{}
}

func (p *AdminAnnouncementsPage) PageTitle() string {
	return "Admin Announcements"
}

var _ tasks.Page = (*AdminAnnouncementsPage)(nil)
var _ tasks.Task = (*AdminAnnouncementsPage)(nil)
var _ http.Handler = (*AdminAnnouncementsPage)(nil)

type AdminCommentsPage struct{}

func (p *AdminCommentsPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminCommentsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Comments"
	queries := cd.Queries()
	rows, err := queries.AdminListAllCommentsWithThreadInfo(r.Context(), db.AdminListAllCommentsWithThreadInfoParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	data := struct {
		*common.CoreData
		Comments []*db.AdminListAllCommentsWithThreadInfoRow
	}{cd, rows}
	AdminCommentsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminCommentsPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Comments", "/admin/comments", &AdminPage{}
}

func (p *AdminCommentsPage) PageTitle() string {
	return "Comments"
}

var _ tasks.Page = (*AdminCommentsPage)(nil)
var _ tasks.Task = (*AdminCommentsPage)(nil)
var _ http.Handler = (*AdminCommentsPage)(nil)

type AdminCommentPage struct {
	CommentID int32
	Data      any
}

func (p *AdminCommentPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Comment %d", p.CommentID), "", &AdminCommentsPage{}
}

func (p *AdminCommentPage) PageTitle() string {
	return fmt.Sprintf("Comment %d", p.CommentID)
}

func (p *AdminCommentPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AdminCommentPageTmpl.Handler(p.Data).ServeHTTP(w, r)
}

var _ tasks.Page = (*AdminCommentPage)(nil)
var _ http.Handler = (*AdminCommentPage)(nil)

type AdminCommentTask struct{}

func (t *AdminCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	c, err := cd.CurrentComment(r)
	if err != nil || c == nil {
		return handlers.ErrNotFound
	}

	queries := cd.Queries()
	rows, err := queries.GetCommentsByIdsForUserWithThreadInfo(r.Context(), db.GetCommentsByIdsForUserWithThreadInfoParams{
		ViewerID: cd.UserID,
		Ids:      []int32{c.Idcomments},
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil || len(rows) == 0 {
		return handlers.ErrNotFound
	}
	comment := rows[0]
	threadRows, _ := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: cd.UserID,
		ThreadID: comment.ForumthreadID,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	var contextRows []*db.GetCommentsByThreadIdForUserRow
	for i, row := range threadRows {
		if row.Idcomments == comment.Idcomments {
			start := i - 3
			if start < 0 {
				start = 0
			}
			end := i + 4
			if end > len(threadRows) {
				end = len(threadRows)
			}
			contextRows = threadRows[start:end]
			break
		}
	}
	data := struct {
		Comment *db.GetCommentsByIdsForUserWithThreadInfoRow
		Context []*db.GetCommentsByThreadIdForUserRow
	}{comment, contextRows}

	return &AdminCommentPage{
		CommentID: c.Idcomments,
		Data: data,
	}
}

type AdminEmailPage struct{}

func (p *AdminEmailPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminEmailPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			handlers.RenderErrorPage(w, r, err)
			return
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
			handlers.RenderErrorPage(w, r, err)
			return
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
		handlers.RenderErrorPage(w, r, err)
		return
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

	AdminEmailPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminEmailPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Email", "/admin/email/queue", &AdminPage{}
}

func (p *AdminEmailPage) PageTitle() string {
	return "Email"
}

var _ tasks.Page = (*AdminEmailPage)(nil)
var _ tasks.Task = (*AdminEmailPage)(nil)
var _ http.Handler = (*AdminEmailPage)(nil)

type AdminUserListPage struct{}

func (p *AdminUserListPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminUserListPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Users"
	if _, err := cd.AdminListUsers(); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	AdminUserListPageTmpl.Handler(struct{}{}).ServeHTTP(w, r)
}

func (p *AdminUserListPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Users", "/admin/user", &AdminPage{}
}

func (p *AdminUserListPage) PageTitle() string {
	return "Users"
}

var _ tasks.Page = (*AdminUserListPage)(nil)
var _ tasks.Task = (*AdminUserListPage)(nil)
var _ http.Handler = (*AdminUserListPage)(nil)

type AdminUserProfilePage struct {
	UserID   int32
	UserName string
	Data     any
}

func (p *AdminUserProfilePage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	title := fmt.Sprintf("User %d", p.UserID)
	if p.UserName != "" {
		title = fmt.Sprintf("User %s", p.UserName)
	}
	return title, "", &AdminUserListPage{}
}

func (p *AdminUserProfilePage) PageTitle() string {
	if p.UserName != "" {
		return fmt.Sprintf("User %s", p.UserName)
	}
	return fmt.Sprintf("User %d", p.UserID)
}

func (p *AdminUserProfilePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AdminUserProfilePageTmpl.Handler(p.Data).ServeHTTP(w, r)
}

var _ tasks.Page = (*AdminUserProfilePage)(nil)
var _ http.Handler = (*AdminUserProfilePage)(nil)

type AdminUserProfileTask struct{}

func (t *AdminUserProfileTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	if user == nil {
		return handlers.ErrNotFound
	}
	return &AdminUserProfilePage{
		UserID: user.Idusers,
		UserName: user.Username.String,
		Data: struct{}{},
	}
}

type AdminSiteSettingsPage struct {
	ConfigFile string
}

func (p *AdminSiteSettingsPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminSiteSettingsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Site Settings"
	cd.FeedsEnabled = cd.Config.FeedsEnabled

	values := config.ValuesMap(*cd.Config)
	defaults := config.DefaultMap(config.NewRuntimeConfig())
	usages := config.UsageMap()
	examples := config.ExamplesMap()
	flags := config.NameMap()

	fileVals, _ := config.LoadAppConfigFile(core.OSFS{}, p.ConfigFile)
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
		ConfigFile: p.ConfigFile,
		Config:     cfg,
	}

	AdminSiteSettingsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminSiteSettingsPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Site Settings", "/admin/settings", &AdminPage{}
}

func (p *AdminSiteSettingsPage) PageTitle() string {
	return "Site Settings"
}

var _ tasks.Page = (*AdminSiteSettingsPage)(nil)
var _ tasks.Task = (*AdminSiteSettingsPage)(nil)
var _ http.Handler = (*AdminSiteSettingsPage)(nil)

type AdminExternalLinksPage struct{}

func (p *AdminExternalLinksPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminExternalLinksPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		handlers.RenderErrorPage(w, r, err)
		return
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
	AdminExternalLinksPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminExternalLinksPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "External Links", "/admin/external-links", &AdminPage{}
}

func (p *AdminExternalLinksPage) PageTitle() string {
	return "External Links"
}

var _ tasks.Page = (*AdminExternalLinksPage)(nil)
var _ tasks.Task = (*AdminExternalLinksPage)(nil)
var _ http.Handler = (*AdminExternalLinksPage)(nil)

type AdminExternalLinkDetailsPage struct {
	LinkID int32
	Data   any
}

func (p *AdminExternalLinkDetailsPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Link %d", p.LinkID), "", &AdminExternalLinksPage{}
}

func (p *AdminExternalLinkDetailsPage) PageTitle() string {
	return fmt.Sprintf("External Link %d", p.LinkID)
}

func (p *AdminExternalLinkDetailsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AdminExternalLinkDetailsPageTmpl.Handler(p.Data).ServeHTTP(w, r)
}

var _ tasks.Page = (*AdminExternalLinkDetailsPage)(nil)
var _ http.Handler = (*AdminExternalLinkDetailsPage)(nil)

type AdminExternalLinkDetailsTask struct{}

func (t *AdminExternalLinkDetailsTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		Link          *db.ExternalLink
		ResultSummary string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasAdminRole() {
		return handlers.ErrForbidden
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return handlers.ErrBadRequest
	}

	queries := cd.Queries()
	link, err := queries.GetExternalLinkByID(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return handlers.ErrNotFound
		}
		return fmt.Errorf("fetching external link: %w", err)
	}

	data := Data{
		Link: link,
	}

	// Result Summary Logic (similar to list page)
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
		data.ResultSummary = fmt.Sprintf("%s external link: %d succeeded, %d failed.", actionLabel, successCount, failureCount)
	}

	return &AdminExternalLinkDetailsPage{
		LinkID: int32(id),
		Data: data,
	}
}
