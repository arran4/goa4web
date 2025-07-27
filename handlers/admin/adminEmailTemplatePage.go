package admin

import (
	"bytes"
	"net/http"
	"sort"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

type taskTemplateInfo struct {
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
	tasksSlice := reg.Registered()
	infos := make([]taskTemplateInfo, 0, len(tasksSlice))
	for _, t := range tasksSlice {
		info := taskTemplateInfo{Task: t.Name()}
		if tp, ok := t.(notif.SelfNotificationTemplateProvider); ok {
			if et := tp.SelfEmailTemplate(); et != nil {
				info.SelfEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt := tp.SelfInternalNotificationTemplate(); nt != nil {
				info.SelfInternal = *nt
			}
		}
		if tp, ok := t.(notif.DirectEmailNotificationTemplateProvider); ok {
			if et := tp.DirectEmailTemplate(); et != nil {
				info.DirectEmail = []string{et.Text, et.HTML, et.Subject}
			}
		}
		if tp, ok := t.(notif.SubscribersNotificationTemplateProvider); ok {
			if et := tp.SubscribedEmailTemplate(); et != nil {
				info.SubEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt := tp.SubscribedInternalNotificationTemplate(); nt != nil {
				info.SubInternal = *nt
			}
		}
		if tp, ok := t.(notif.AdminEmailTemplateProvider); ok {
			if et := tp.AdminEmailTemplate(); et != nil {
				info.AdminEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt := tp.AdminInternalNotificationTemplate(); nt != nil {
				info.AdminInternal = *nt
			}
		}
		if tp, ok := t.(notif.TargetUsersNotificationProvider); ok {
			if et := tp.TargetEmailTemplate(); et != nil {
				info.TargetEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt := tp.TargetInternalNotificationTemplate(); nt != nil {
				info.TargetInternal = *nt
			}
		}
		infos = append(infos, info)
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Task < infos[j].Task })
	return infos
}

func defaultTemplate(name string, cfg *config.RuntimeConfig) string {
	var buf bytes.Buffer
	if strings.HasSuffix(name, ".gohtml") {
		tmpl := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
		if err := tmpl.ExecuteTemplate(&buf, name, sampleEmailData(cfg)); err == nil {
			return buf.String()
		}
	} else {
		tmpl := templates.GetCompiledEmailTextTemplates(map[string]any{})
		if err := tmpl.ExecuteTemplate(&buf, name, sampleEmailData(cfg)); err == nil {
			return buf.String()
		}
		tmpl2 := templates.GetCompiledNotificationTemplates(map[string]any{})
		buf.Reset()
		if err := tmpl2.ExecuteTemplate(&buf, name, sampleEmailData(cfg)); err == nil {
			return buf.String()
		}
	}
	return ""
}

// AdminEmailTemplatePage provides template listing and editing.
func AdminEmailTemplatePage(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if name == "" {
		data := struct {
			*common.CoreData
			Infos []taskTemplateInfo
		}{cd, gatherTaskTemplateInfos(cd.TasksReg)}
		handlers.TemplateHandler(w, r, "emailTemplateListPage.gohtml", data)
		return
	}
	q := cd.Queries()
	body, _ := q.GetTemplateOverride(r.Context(), name)
	data := struct {
		*common.CoreData
		Name    string
		Body    string
		Default string
		Error   string
	}{
		CoreData: cd,
		Name:     name,
		Body:     body,
		Default:  defaultTemplate(name, cd.Config),
		Error:    r.URL.Query().Get("error"),
	}
	handlers.TemplateHandler(w, r, "emailTemplateEditPage.gohtml", data)
}

func sampleEmailData(cfg *config.RuntimeConfig) map[string]interface{} {
	return map[string]interface{}{
		"URL":            "http://example.com",
		"UnsubscribeUrl": "http://example.com/unsub",
		"From":           cfg.EmailFrom,
		"To":             "user@example.com",
	}
}
