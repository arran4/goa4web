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

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// ClearExpiredPasswordResetsTask removes expired or verified password reset requests.
type ClearExpiredPasswordResetsTask struct{ tasks.TaskString }

var clearExpiredPasswordResetsTask = &ClearExpiredPasswordResetsTask{TaskString: TaskClearExpiredPasswordResets}

var _ tasks.Task = (*ClearExpiredPasswordResetsTask)(nil)
var _ tasks.AuditableTask = (*ClearExpiredPasswordResetsTask)(nil)

func (ClearExpiredPasswordResetsTask) Action(_ http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	hours, err := parseCleanupHours(r.FormValue("hours"))
	if err != nil {
		return fmt.Errorf("invalid hours: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	expiry := time.Now().Add(-time.Duration(hours) * time.Hour)
	res, err := queries.SystemPurgePasswordResetsBefore(r.Context(), expiry)
	if err != nil {
		return fmt.Errorf("clear expired resets: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	removed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["DeletedCount"] = removed
		evt.Data["Hours"] = hours
	}

	return handlers.RedirectHandler(passwordResetCleanupRedirect(r, url.Values{
		"cleanup_action": []string{"expired"},
		"cleanup_count":  []string{strconv.FormatInt(removed, 10)},
		"cleanup_hours":  []string{strconv.Itoa(hours)},
	}))
}

// AuditRecord summarises clearing expired password reset requests.
func (ClearExpiredPasswordResetsTask) AuditRecord(data map[string]any) string {
	count := auditCount(data["DeletedCount"])
	if hours, ok := data["Hours"].(int); ok {
		return fmt.Sprintf("cleared %d expired password reset requests older than %d hours", count, hours)
	}
	return fmt.Sprintf("cleared %d expired password reset requests", count)
}

// ClearUserPasswordResetsTask removes password reset requests for a user.
type ClearUserPasswordResetsTask struct{ tasks.TaskString }

var clearUserPasswordResetsTask = &ClearUserPasswordResetsTask{TaskString: TaskClearUserPasswordResets}

var _ tasks.Task = (*ClearUserPasswordResetsTask)(nil)
var _ tasks.AuditableTask = (*ClearUserPasswordResetsTask)(nil)

func (ClearUserPasswordResetsTask) Action(_ http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	userRaw := strings.TrimSpace(r.FormValue("user"))
	if userRaw == "" {
		return fmt.Errorf("user required: %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("missing user")))
	}

	userID, username, err := resolveCleanupUser(r.Context(), queries, userRaw)
	if err != nil {
		return fmt.Errorf("resolve user: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	res, err := queries.SystemDeletePasswordResetsByUser(r.Context(), userID)
	if err != nil {
		return fmt.Errorf("clear user resets: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	removed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["DeletedCount"] = removed
		evt.Data["UserID"] = userID
		evt.Data["Username"] = username
	}

	return handlers.RedirectHandler(passwordResetCleanupRedirect(r, url.Values{
		"cleanup_action": []string{"user"},
		"cleanup_count":  []string{strconv.FormatInt(removed, 10)},
		"cleanup_user":   []string{username},
	}))
}

// AuditRecord summarises clearing password reset requests for a user.
func (ClearUserPasswordResetsTask) AuditRecord(data map[string]any) string {
	count := auditCount(data["DeletedCount"])
	if username, ok := data["Username"].(string); ok {
		return fmt.Sprintf("cleared %d password reset requests for %s", count, username)
	}
	if id, ok := data["UserID"].(int32); ok {
		return fmt.Sprintf("cleared %d password reset requests for user %d", count, id)
	}
	return fmt.Sprintf("cleared %d password reset requests for user", count)
}

func parseCleanupHours(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 24, nil
	}
	hours, err := strconv.Atoi(raw)
	if err != nil || hours < 1 {
		return 0, fmt.Errorf("invalid hours %q", raw)
	}
	return hours, nil
}

func resolveCleanupUser(ctx context.Context, queries db.Querier, raw string) (int32, string, error) {
	if id, err := strconv.Atoi(raw); err == nil {
		if id < 1 {
			return 0, "", fmt.Errorf("invalid user id: %s", raw)
		}
		user, err := queries.SystemGetUserByID(ctx, int32(id))
		if err != nil {
			return 0, "", err
		}
		return user.Idusers, user.Username.String, nil
	}
	user, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: raw, Valid: true})
	if err != nil {
		return 0, "", err
	}
	return user.Idusers, user.Username.String, nil
}

func passwordResetCleanupRedirect(r *http.Request, params url.Values) string {
	back := strings.TrimSpace(r.FormValue("back"))
	target, err := url.Parse(back)
	if err != nil || target.Path == "" || !strings.HasPrefix(target.Path, "/admin/password_resets") {
		target = &url.URL{Path: "/admin/password_resets"}
	}
	values := target.Query()
	for key, vals := range params {
		if len(vals) > 0 {
			values.Set(key, vals[0])
		}
	}
	target.RawQuery = values.Encode()
	return target.String()
}

func auditCount(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case uint:
		return int64(v)
	default:
		return 0
	}
}
