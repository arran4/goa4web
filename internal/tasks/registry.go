package tasks

import "sync"

// NamedTask identifies a task by name.
type NamedTask interface {
	Name() string
}

// Registry stores registered tasks.
type Registry struct {
	mu    sync.Mutex
	tasks []NamedTask
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{} }

// Register adds t to the Registry. Duplicate names are ignored.
func (r *Registry) Register(t NamedTask) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rt := range r.tasks {
		if rt.Name() == t.Name() {
			return
		}
	}
	r.tasks = append(r.tasks, t)
}

// Registered returns a copy of the registered tasks slice.
func (r *Registry) Registered() []NamedTask {
	r.mu.Lock()
	tasks := append([]NamedTask(nil), r.tasks...)
	r.mu.Unlock()
	return tasks
}
