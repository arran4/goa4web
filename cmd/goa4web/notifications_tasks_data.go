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
		if et, nt, ok := notif.SelfTemplates(t, evt); ok {
			if et != nil {
				info.SelfEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt != nil {
				info.SelfInternal = *nt
			}
		}
		if et, _, ok, _ := notif.DirectEmailTemplate(t, evt); ok {
			if et != nil {
				info.DirectEmail = []string{et.Text, et.HTML, et.Subject}
			}
		}
		if et, nt, ok := notif.SubscriberTemplates(t, evt); ok {
			if et != nil {
				info.SubEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt != nil {
				info.SubInternal = *nt
			}
		}
		if et, nt, ok := notif.AdminTemplates(t, evt); ok {
			if et != nil {
				info.AdminEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt != nil {
				info.AdminInternal = *nt
			}
		}
		if _, et, nt, ok, _ := notif.TargetUsersTemplates(t, evt); ok {
			if et != nil {
				info.TargetEmail = []string{et.Text, et.HTML, et.Subject}
			}
			if nt != nil {
				info.TargetInternal = *nt
			}
		}
		infos = append(infos, info)
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Task < infos[j].Task })
	return infos
}
