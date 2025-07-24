package main

import (
	"sort"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// taskTemplateInfo describes the template files used for a task.
type taskTemplateInfo struct {
	Task          string
	SelfEmail     []string
	SelfInternal  string
	SubEmail      []string
	SubInternal   string
	AdminEmail    []string
	AdminInternal string
}

func taskTemplateInfos(reg *tasks.Registry) []taskTemplateInfo {
	tasks := reg.Registered()
	infos := make([]taskTemplateInfo, 0, len(tasks))
	for _, t := range tasks {
		info := taskTemplateInfo{Task: t.Name()}
		if tp, ok := t.(notif.SelfNotificationTemplateProvider); ok {
			if et := tp.SelfEmailTemplate(); et != nil {
				info.SelfEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt := tp.SelfInternalNotificationTemplate(); nt != nil {
				info.SelfInternal = *nt
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
		infos = append(infos, info)
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Task < infos[j].Task })
	return infos
}
