package handlers

import (
	"reflect"
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// TaskRegistry holds all tasks that should be tested for notification templates.
type TaskRegistry struct {
	tasks []TaskInfo
}

// TaskInfo describes a task and how to create a test event for it.
type TaskInfo struct {
	Name         string
	Task         tasks.Task
	EventFactory func() eventbus.TaskEvent
}

// NewTaskRegistry creates a new registry for testing notification templates.
func NewTaskRegistry() *TaskRegistry {
	return &TaskRegistry{
		tasks: make([]TaskInfo, 0),
	}
}

// Register adds a task to the registry with its event factory.
func (r *TaskRegistry) Register(name string, task tasks.Task, eventFactory func() eventbus.TaskEvent) {
	r.tasks = append(r.tasks, TaskInfo{
		Name:         name,
		Task:         task,
		EventFactory: eventFactory,
	})
}

// TestAll runs notification template tests for all registered tasks.
func (r *TaskRegistry) TestAll(t *testing.T) {
	for _, info := range r.tasks {
		t.Run(info.Name, func(t *testing.T) {
			evt := info.EventFactory()
			TestNotificationTemplates(t, []NotificationTemplateTest{
				{
					Name:  info.Name,
					Task:  info.Task,
					Event: evt,
				},
			})
		})
	}
}

// AutoDiscoverTasks automatically discovers tasks that implement notification interfaces.
// This is useful for ensuring all tasks are tested.
func AutoDiscoverTasks(t *testing.T, taskInstances ...tasks.Task) []TaskInfo {
	t.Helper()

	var discovered []TaskInfo

	for _, task := range taskInstances {
		name := reflect.TypeOf(task).String()

		// Check if it implements any notification interface
		hasNotifications := false
		if _, ok := task.(notif.AdminEmailTemplateProvider); ok {
			hasNotifications = true
		}
		if _, ok := task.(notif.SelfNotificationTemplateProvider); ok {
			hasNotifications = true
		}
		if _, ok := task.(notif.TargetUsersNotificationProvider); ok {
			hasNotifications = true
		}
		if _, ok := task.(notif.SubscribersNotificationTemplateProvider); ok {
			hasNotifications = true
		}
		if _, ok := task.(notif.DirectEmailNotificationTemplateProvider); ok {
			hasNotifications = true
		}

		if hasNotifications {
			discovered = append(discovered, TaskInfo{
				Name: name,
				Task: task,
				EventFactory: func() eventbus.TaskEvent {
					return CreateTestEvent(task, "/test", 1, map[string]any{
						"Username": "testuser",
					})
				},
			})
		}
	}

	return discovered
}

// AssertNoMissingNotificationTests checks that all notification provider tasks are tested.
// This helps prevent forgetting to add tests when new tasks are created.
func AssertNoMissingNotificationTests(t *testing.T, registry *TaskRegistry, allTasks []tasks.Task) {
	t.Helper()

	registered := make(map[string]bool)
	for _, info := range registry.tasks {
		registered[info.Name] = true
	}

	discovered := AutoDiscoverTasks(t, allTasks...)

	var missing []string
	for _, info := range discovered {
		if !registered[info.Name] {
			missing = append(missing, info.Name)
		}
	}

	if len(missing) > 0 {
		t.Errorf("Found %d tasks with notification templates but no tests: %v", len(missing), missing)
	}
}
