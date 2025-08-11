package main

import (
	"sort"

	"github.com/arran4/goa4web/internal/eventbus"
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
