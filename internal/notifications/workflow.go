package notifications

import (
	"sync"

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

type Workflow struct {
	AdminEmail       func(eventbus.TaskEvent) (*EmailTemplates, *string, bool)
	SelfNotify       func(eventbus.TaskEvent) (*EmailTemplates, *string, bool)
	DirectEmail      func(eventbus.TaskEvent) (*EmailTemplates, string, bool, error)
	TargetUsers      func(eventbus.TaskEvent) ([]int32, *EmailTemplates, *string, bool, error)
	SubscriberNotify func(eventbus.TaskEvent) (*EmailTemplates, *string, bool)
	AutoSubscribe    func(eventbus.TaskEvent) (string, string, []GrantRequirement, bool, error)
	Grants           func(eventbus.TaskEvent) ([]GrantRequirement, bool, error)
	SelfBroadcast    func() bool
}

func (w Workflow) HasAdmin() bool         { return w.AdminEmail != nil }
func (w Workflow) HasSelf() bool          { return w.SelfNotify != nil }
func (w Workflow) HasDirect() bool        { return w.DirectEmail != nil }
func (w Workflow) HasTargetUsers() bool   { return w.TargetUsers != nil }
func (w Workflow) HasSubscriber() bool    { return w.SubscriberNotify != nil }
func (w Workflow) HasAutoSubscribe() bool { return w.AutoSubscribe != nil }

type workflowRegistry struct {
	mu        sync.RWMutex
	workflows map[string]Workflow
}

var defaultWorkflowRegistry = &workflowRegistry{workflows: map[string]Workflow{}}

// RegisterWorkflow lodges explicit notification behavior for a task name.
func RegisterWorkflow(taskName string, wf Workflow) {
	defaultWorkflowRegistry.mu.Lock()
	defer defaultWorkflowRegistry.mu.Unlock()
	defaultWorkflowRegistry.workflows[taskName] = wf
}

func (r *workflowRegistry) lookup(taskName string) (Workflow, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wf, ok := r.workflows[taskName]
	return wf, ok
}

// WorkflowForTask returns the effective workflow for task.
// Explicitly registered workflows are preferred over method-derived behavior.
func WorkflowForTask(task any) Workflow {
	name := ""
	if n, ok := task.(tasks.Name); ok {
		name = n.Name()
	}
	if name != "" {
		if wf, ok := defaultWorkflowRegistry.lookup(name); ok {
			return wf
		}
	}
	return deriveWorkflow(task)
}

func deriveWorkflow(task any) Workflow {
	wf := Workflow{}
	if task == nil {
		return wf
	}

	if t, ok := task.(interface {
		AdminEmailTemplate(eventbus.TaskEvent) (*EmailTemplates, bool)
		AdminInternalNotificationTemplate(eventbus.TaskEvent) *string
	}); ok {
		wf.AdminEmail = func(evt eventbus.TaskEvent) (*EmailTemplates, *string, bool) {
			et, send := t.AdminEmailTemplate(evt)
			if !send {
				return nil, nil, false
			}
			return et, t.AdminInternalNotificationTemplate(evt), true
		}
	}

	if t, ok := task.(interface {
		SelfEmailTemplate(eventbus.TaskEvent) (*EmailTemplates, bool)
		SelfInternalNotificationTemplate(eventbus.TaskEvent) *string
	}); ok {
		wf.SelfNotify = func(evt eventbus.TaskEvent) (*EmailTemplates, *string, bool) {
			et, send := t.SelfEmailTemplate(evt)
			if !send {
				et = nil
			}
			nt := t.SelfInternalNotificationTemplate(evt)
			return et, nt, et != nil || nt != nil
		}
	}

	if t, ok := task.(interface{ SelfEmailBroadcast() bool }); ok {
		wf.SelfBroadcast = t.SelfEmailBroadcast
	}

	if t, ok := task.(interface {
		DirectEmailAddress(eventbus.TaskEvent) (string, error)
		DirectEmailTemplate(eventbus.TaskEvent) (*EmailTemplates, bool)
	}); ok {
		wf.DirectEmail = func(evt eventbus.TaskEvent) (*EmailTemplates, string, bool, error) {
			et, send := t.DirectEmailTemplate(evt)
			if !send {
				return nil, "", false, nil
			}
			addr, err := t.DirectEmailAddress(evt)
			if err != nil {
				return nil, "", false, err
			}
			return et, addr, true, nil
		}
	}

	if t, ok := task.(interface {
		TargetUserIDs(eventbus.TaskEvent) ([]int32, error)
		TargetEmailTemplate(eventbus.TaskEvent) (*EmailTemplates, bool)
		TargetInternalNotificationTemplate(eventbus.TaskEvent) *string
	}); ok {
		wf.TargetUsers = func(evt eventbus.TaskEvent) ([]int32, *EmailTemplates, *string, bool, error) {
			ids, err := t.TargetUserIDs(evt)
			if err != nil {
				return nil, nil, nil, false, err
			}
			et, _ := t.TargetEmailTemplate(evt)
			return ids, et, t.TargetInternalNotificationTemplate(evt), true, nil
		}
	}

	if t, ok := task.(interface {
		SubscribedEmailTemplate(eventbus.TaskEvent) (*EmailTemplates, bool)
		SubscribedInternalNotificationTemplate(eventbus.TaskEvent) *string
	}); ok {
		wf.SubscriberNotify = func(evt eventbus.TaskEvent) (*EmailTemplates, *string, bool) {
			et, send := t.SubscribedEmailTemplate(evt)
			if !send {
				et = nil
			}
			return et, t.SubscribedInternalNotificationTemplate(evt), true
		}
	}

	if t, ok := task.(interface {
		AutoSubscribePath(eventbus.TaskEvent) (string, string, error)
		AutoSubscribeGrants(eventbus.TaskEvent) ([]GrantRequirement, error)
	}); ok {
		wf.AutoSubscribe = func(evt eventbus.TaskEvent) (string, string, []GrantRequirement, bool, error) {
			action, path, err := t.AutoSubscribePath(evt)
			if err != nil {
				return "", "", nil, true, err
			}
			grants, err := t.AutoSubscribeGrants(evt)
			return action, path, grants, true, err
		}
	}

	if t, ok := task.(interface {
		GrantsRequired(eventbus.TaskEvent) ([]GrantRequirement, error)
	}); ok {
		wf.Grants = func(evt eventbus.TaskEvent) ([]GrantRequirement, bool, error) {
			grants, err := t.GrantsRequired(evt)
			return grants, true, err
		}
	}

	return wf
}
