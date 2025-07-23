package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/common"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AddIPBanTask blocks a network from accessing the site.
type AddIPBanTask struct{ tasks.TaskString }

var addIPBanTask = &AddIPBanTask{TaskString: TaskAdd}

// DeleteIPBanTask lifts an IP ban.
type DeleteIPBanTask struct{ tasks.TaskString }

var deleteIPBanTask = &DeleteIPBanTask{TaskString: TaskDelete}

var _ tasks.Task = (*AddIPBanTask)(nil)
var _ tasks.AuditableTask = (*AddIPBanTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AddIPBanTask)(nil)

var _ tasks.Task = (*DeleteIPBanTask)(nil)
var _ tasks.AuditableTask = (*DeleteIPBanTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*DeleteIPBanTask)(nil)

func AdminIPBanPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Bans []*db.BannedIp
	}
	data := Data{CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	rows, err := queries.ListBannedIps(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list banned ips: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Bans = rows
	handlers.TemplateHandler(w, r, "iPBanPage.gohtml", data)
}

func (AddIPBanTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

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
		if err := queries.InsertBannedIp(r.Context(), db.InsertBannedIpParams{
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
		if u, _ := cd.CurrentUser(); u != nil {
			evt.Data["Moderator"] = u.Username
		}
	}
	return nil
}

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

func (AddIPBanTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminAddIPBanEmail")
}

func (AddIPBanTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminAddIPBanEmail")
	return &v
}

func (DeleteIPBanTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminRemoveIPBanEmail")
}

func (DeleteIPBanTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminRemoveIPBanEmail")
	return &v
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

// AuditRecord summarises the removal of an IP ban.
func (DeleteIPBanTask) AuditRecord(data map[string]any) string {
	ip, _ := data["IP"].(string)
	mod, _ := data["Moderator"].(string)
	return fmt.Sprintf("%s removed ban on %s", mod, ip)
}
