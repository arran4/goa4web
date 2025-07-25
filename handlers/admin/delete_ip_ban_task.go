package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteIPBanTask lifts an IP ban.
type DeleteIPBanTask struct{ tasks.TaskString }

var deleteIPBanTask = &DeleteIPBanTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteIPBanTask)(nil)
var _ tasks.AuditableTask = (*DeleteIPBanTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*DeleteIPBanTask)(nil)

func (DeleteIPBanTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	var ips []string
	for _, ip := range r.Form["ip"] {
		ipNet := NormalizeIPNet(ip)
		if err := queries.CancelBannedIp(r.Context(), ipNet); err != nil {
			return fmt.Errorf("cancel banned ip %s fail %w", ipNet, handlers.ErrRedirectOnSamePageHandler(err))
		}
		if ipNet != "" {
			ips = append(ips, ipNet)
		}
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["IP"] = strings.Join(ips, ", ")
		if u, _ := cd.CurrentUser(); u != nil {
			evt.Data["Moderator"] = u.Username
		}
	}
	return nil
}

func (DeleteIPBanTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminRemoveIPBanEmail")
}

func (DeleteIPBanTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminRemoveIPBanEmail")
	return &v
}

// AuditRecord summarises the removal of an IP ban.
func (DeleteIPBanTask) AuditRecord(data map[string]any) string {
	ip, _ := data["IP"].(string)
	mod, _ := data["Moderator"].(string)
	return fmt.Sprintf("%s removed ban on %s", mod, ip)
}
