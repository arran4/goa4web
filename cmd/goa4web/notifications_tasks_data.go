package main

import (
	"sort"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// taskTemplateInfo describes the template files used for a task.
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

func taskTemplateInfos(reg *tasks.Registry) []taskTemplateInfo {
	entries := reg.Entries()
	infos := make([]taskTemplateInfo, 0, len(entries))
	for _, e := range entries {
		info := taskTemplateInfo{Section: e.Section, Task: e.Task.Name()}
		t := e.Task
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
