package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

const AdminPasswordResetListPageTmpl tasks.Template = "admin/passwordResetList.gohtml"

type AdminPasswordResetListPageData struct {
	Rows           []*db.AdminListPasswordResetsRow
	Status         string
	UserFilter     string
	AgeHours       string
	CleanupUser    string
	CleanupAge     int
	Page           int
	TotalPages     int
	FiltersQuery   string
	ReturnURL      string
	SummaryMessage string
}

func adminPasswordResetListPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Password Resets"

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		idStr := r.FormValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid id: %w", err))
			return
		}

		switch action {
		case "approve":
			if err := cd.AdminApprovePasswordReset(int32(id)); err != nil {
				handlers.RenderErrorPage(w, r, fmt.Errorf("approve failed: %w", err))
				return
			}
		case "deny":
			if err := cd.AdminDenyPasswordReset(int32(id)); err != nil {
				handlers.RenderErrorPage(w, r, fmt.Errorf("deny failed: %w", err))
				return
			}
		}
		http.Redirect(w, r, r.URL.Path+"?"+r.URL.RawQuery, http.StatusSeeOther)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		status = "pending"
	}
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	userFilter := strings.TrimSpace(r.FormValue("user"))
	ageHoursRaw := strings.TrimSpace(r.FormValue("age_hours"))
	var ageHours int
	if ageHoursRaw != "" {
		parsed, err := strconv.Atoi(ageHoursRaw)
		if err != nil || parsed < 0 {
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid age hours: %s", ageHoursRaw))
			return
		}
		ageHours = parsed
	}

	var userID *int32
	if userFilter != "" {
		id, err := resolvePasswordResetUser(r.Context(), cd.Queries(), userFilter)
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		userID = &id
	}

	pageStr := r.FormValue("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	const pageSize = 20
	offset := int32((page - 1) * pageSize)

	var createdBefore *time.Time
	if ageHours > 0 {
		cutoff := time.Now().Add(-time.Duration(ageHours) * time.Hour)
		createdBefore = &cutoff
	}

	rows, count, err := cd.AdminListPasswordResets(statusPtr, userID, createdBefore, pageSize, offset)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("list failed: %w", err))
		return
	}

	totalPages := int((count + pageSize - 1) / pageSize)
	filtersQuery := buildPasswordResetFiltersQuery(status, userFilter, ageHoursRaw)
	returnURL := r.URL.Path
	if filtersQuery != "" {
		returnURL = returnURL + "?" + filtersQuery
	}

	summaryMessage := passwordResetSummaryMessage(r.URL.Query())

	AdminPasswordResetListPageTmpl.Handle(w, r, &AdminPasswordResetListPageData{
		Rows:           rows,
		Status:         status,
		UserFilter:     userFilter,
		AgeHours:       ageHoursRaw,
		CleanupUser:    userFilter,
		CleanupAge:     defaultCleanupAge(ageHours),
		Page:           page,
		TotalPages:     totalPages,
		FiltersQuery:   filtersQuery,
		ReturnURL:      returnURL,
		SummaryMessage: summaryMessage,
	})
}

func resolvePasswordResetUser(ctx context.Context, queries db.Querier, raw string) (int32, error) {
	if id, err := strconv.Atoi(raw); err == nil {
		if id < 1 {
			return 0, fmt.Errorf("invalid user id: %s", raw)
		}
		if _, err := queries.SystemGetUserByID(ctx, int32(id)); err != nil {
			return 0, fmt.Errorf("find user id %d: %w", id, err)
		}
		return int32(id), nil
	}
	user, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: raw, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("find username %s: %w", raw, err)
	}
	return user.Idusers, nil
}

func buildPasswordResetFiltersQuery(status, user, ageHours string) string {
	values := url.Values{}
	if status != "" {
		values.Set("status", status)
	}
	if user != "" {
		values.Set("user", user)
	}
	if ageHours != "" {
		values.Set("age_hours", ageHours)
	}
	return values.Encode()
}

func defaultCleanupAge(ageHours int) int {
	if ageHours > 0 {
		return ageHours
	}
	return 24
}

func passwordResetSummaryMessage(values url.Values) string {
	action := values.Get("cleanup_action")
	if action == "" {
		return ""
	}
	count, _ := strconv.Atoi(values.Get("cleanup_count"))
	switch action {
	case "expired":
		hours := values.Get("cleanup_hours")
		if hours != "" {
			return fmt.Sprintf("Removed %d password reset requests older than %s hours (including verified entries).", count, hours)
		}
		return fmt.Sprintf("Removed %d expired or verified password reset requests.", count)
	case "user":
		user := values.Get("cleanup_user")
		if user != "" {
			return fmt.Sprintf("Removed %d password reset requests for %s.", count, user)
		}
		return fmt.Sprintf("Removed %d password reset requests for the selected user.", count)
	default:
		return ""
	}
}
