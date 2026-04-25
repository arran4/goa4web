package notifications

import (
	"github.com/arran4/goa4web/internal/eventbus"
)

// AdminTemplates returns admin email/internal notification templates for task.
func AdminTemplates(task any, evt eventbus.TaskEvent) (*EmailTemplates, *string, bool) {
	wf := WorkflowForTask(task)
	if wf.AdminEmail == nil {
		return nil, nil, false
	}
	return wf.AdminEmail(evt)
}

// SelfTemplates returns self email/internal notification templates for task.
func SelfTemplates(task any, evt eventbus.TaskEvent) (*EmailTemplates, *string, bool) {
	wf := WorkflowForTask(task)
	if wf.SelfNotify == nil {
		return nil, nil, false
	}
	return wf.SelfNotify(evt)
}

// DirectEmailTemplate resolves direct email template and target address.
func DirectEmailTemplate(task any, evt eventbus.TaskEvent) (*EmailTemplates, string, bool, error) {
	wf := WorkflowForTask(task)
	if wf.DirectEmail == nil {
		return nil, "", false, nil
	}
	return wf.DirectEmail(evt)
}

// SubscriberTemplates returns subscriber email/internal notification templates.
func SubscriberTemplates(task any, evt eventbus.TaskEvent) (*EmailTemplates, *string, bool) {
	wf := WorkflowForTask(task)
	if wf.SubscriberNotify == nil {
		return nil, nil, false
	}
	return wf.SubscriberNotify(evt)
}

// TargetUsersTemplates returns target-user notification data.
func TargetUsersTemplates(task any, evt eventbus.TaskEvent) ([]int32, *EmailTemplates, *string, bool, error) {
	wf := WorkflowForTask(task)
	if wf.TargetUsers == nil {
		return nil, nil, nil, false, nil
	}
	return wf.TargetUsers(evt)
}

// AutoSubscribeData returns auto-subscribe action/path/grants for task.
func AutoSubscribeData(task any, evt eventbus.TaskEvent) (string, string, []GrantRequirement, bool, error) {
	wf := WorkflowForTask(task)
	if wf.AutoSubscribe == nil {
		return "", "", nil, false, nil
	}
	return wf.AutoSubscribe(evt)
}

// HasAutoSubscribe reports whether task supports auto-subscribe methods.
func HasAutoSubscribe(task any) bool {
	return WorkflowForTask(task).HasAutoSubscribe()
}

// HasAdminTemplates reports whether task exposes admin notification templates.
func HasAdminTemplates(task any) bool {
	return WorkflowForTask(task).HasAdmin()
}

// HasSelfTemplates reports whether task exposes self notification templates.
func HasSelfTemplates(task any) bool {
	return WorkflowForTask(task).HasSelf()
}

// HasTargetUsersTemplates reports whether task exposes target-user templates.
func HasTargetUsersTemplates(task any) bool {
	return WorkflowForTask(task).HasTargetUsers()
}

// HasSubscriberTemplates reports whether task exposes subscriber templates.
func HasSubscriberTemplates(task any) bool {
	return WorkflowForTask(task).HasSubscriber()
}

// HasDirectEmailTemplate reports whether task exposes a direct-email template.
func HasDirectEmailTemplate(task any) bool {
	return WorkflowForTask(task).HasDirect()
}
