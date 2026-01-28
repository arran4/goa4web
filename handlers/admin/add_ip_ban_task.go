package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// AddIPBanTask blocks a network from accessing the site.
type AddIPBanTask struct{ tasks.TaskString }

var addIPBanTask = &AddIPBanTask{TaskString: TaskAdd}

var _ tasks.Task = (*AddIPBanTask)(nil)
var _ tasks.AuditableTask = (*AddIPBanTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AddIPBanTask)(nil)
var _ tasks.EmailTemplatesRequired = (*AddIPBanTask)(nil)

func (AddIPBanTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { handlers.RenderErrorPage(w, r, handlers.ErrForbidden) })
	}
	queries := cd.Queries()

	ipNet := strings.TrimSpace(r.PostFormValue("ip"))
	ipNet = NormalizeIPNet(ipNet)
	reason := strings.TrimSpace(r.PostFormValue("reason"))
	expiresStr := strings.TrimSpace(r.PostFormValue("expires"))
	var expires sql.NullTime
	if expiresStr != "" {
		if t, err := time.Parse("2006-01-02", expiresStr); err == nil {
			expires = sql.NullTime{Time: t, Valid: true}
		}
	}
	if ipNet != "" {
		if err := queries.AdminInsertBannedIp(r.Context(), db.AdminInsertBannedIpParams{
			IpNet:     ipNet,
			Reason:    sql.NullString{String: reason, Valid: reason != ""},
			ExpiresAt: expires,
		}); err != nil {
			return fmt.Errorf("insert banned ip fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["IP"] = ipNet
		if reason != "" {
			evt.Data["Reason"] = reason
		}
		if u, _ := cd.CurrentUser(); u != nil && u.Username.Valid {
			evt.Data["Moderator"] = u.Username.String
		}
	}
	return nil
}

func (AddIPBanTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminAddIPBan.EmailTemplates(), true
}

func (AddIPBanTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminAddIPBan.NotificationTemplate()
	return &v
}

func (AddIPBanTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateAdminAddIPBan.RequiredTemplates()
}

// AuditRecord summarises the addition of an IP ban.
func (AddIPBanTask) AuditRecord(data map[string]any) string {
	ip, _ := data["IP"].(string)
	mod, _ := data["Moderator"].(string)
	reason, _ := data["Reason"].(string)
	if reason != "" {
		return fmt.Sprintf("%s banned %s (%s)", mod, ip, reason)
	}
	return fmt.Sprintf("%s banned %s", mod, ip)
}
