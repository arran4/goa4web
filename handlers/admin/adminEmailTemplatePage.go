package admin

import (
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

type taskTemplateInfo struct {
	Section        string
	Task           string
	SelfEmail      []string
	SelfInternal   string
	DirectEmail    []string
	SubEmail       []string
	SubInternal    string
	TargetEmail    []string
	TargetInternal string
	AdminEmail     []string
	AdminInternal  string
}

func gatherTaskTemplateInfos(reg *tasks.Registry) []taskTemplateInfo {
	entries := reg.Entries()
	infos := make([]taskTemplateInfo, 0, len(entries))
	for _, e := range entries {
		info := taskTemplateInfo{Section: e.Section, Task: e.Task.Name()}
		t := e.Task
		evt := eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}
		if tp, ok := t.(notif.SelfNotificationTemplateProvider); ok {
			if et, _ := tp.SelfEmailTemplate(evt); et != nil {
				info.SelfEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt := tp.SelfInternalNotificationTemplate(evt); nt != nil {
				info.SelfInternal = *nt
			}
		}
		if tp, ok := t.(notif.DirectEmailNotificationTemplateProvider); ok {
			if et, _ := tp.DirectEmailTemplate(evt); et != nil {
				info.DirectEmail = []string{et.Text, et.HTML, et.Subject}
			}
		}
		if tp, ok := t.(notif.SubscribersNotificationTemplateProvider); ok {
			if et, _ := tp.SubscribedEmailTemplate(evt); et != nil {
				info.SubEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt := tp.SubscribedInternalNotificationTemplate(evt); nt != nil {
				info.SubInternal = *nt
			}
		}
		if tp, ok := t.(notif.AdminEmailTemplateProvider); ok {
			if et, _ := tp.AdminEmailTemplate(evt); et != nil {
				info.AdminEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt := tp.AdminInternalNotificationTemplate(evt); nt != nil {
				info.AdminInternal = *nt
			}
		}
		if tp, ok := t.(notif.TargetUsersNotificationProvider); ok {
			if et, _ := tp.TargetEmailTemplate(evt); et != nil {
				info.TargetEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt := tp.TargetInternalNotificationTemplate(evt); nt != nil {
				info.TargetInternal = *nt
			}
		}
		infos = append(infos, info)
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Task < infos[j].Task })
	return infos
}

// AdminEmailTemplatePage provides template listing and editing.
func AdminEmailTemplatePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Email Templates"
	name := r.URL.Query().Get("name")
	if name == "" {
		data := struct {
			*common.CoreData
			Infos []taskTemplateInfo
		}{cd, gatherTaskTemplateInfos(cd.TasksReg)}
		AdminEmailTemplateListPageTmpl.Handle(w, r, data)
		return
	}
	errMsg := r.URL.Query().Get("error")
	cd.SetCurrentNotificationTemplate(name, errMsg)
	cd.SetCurrentError(errMsg)
	AdminEmailTemplateEditPageTmpl.Handle(w, r, struct{}{})
}

const AdminEmailTemplateListPageTmpl handlers.Page = "admin/emailTemplateListPage.gohtml"

const AdminEmailTemplateEditPageTmpl handlers.Page = "admin/emailTemplateEditPage.gohtml"
