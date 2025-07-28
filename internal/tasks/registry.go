package tasks

import "sync"

// NamedTask identifies a task by name.
type NamedTask interface {
	Name() string
}

// Entry ties a task to a section.
type Entry struct {
	Section string
	Task    NamedTask
}

// Registry stores registered tasks.
type Registry struct {
	mu    sync.Mutex
	tasks []Entry
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{} }

// Register adds t to the Registry under section. Duplicate section/name pairs are ignored.
func (r *Registry) Register(section string, t NamedTask) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range r.tasks {
		if e.Section == section && e.Task.Name() == t.Name() {
			return
		}
	}
	r.tasks = append(r.tasks, Entry{Section: section, Task: t})
}

// Registered returns all registered tasks without section information.
func (r *Registry) Registered() []NamedTask {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]NamedTask, len(r.tasks))
	for i, e := range r.tasks {
		out[i] = e.Task
	}
	return out
}

// Entries returns the registered tasks including section details.
func (r *Registry) Entries() []Entry {
	r.mu.Lock()
	entries := append([]Entry(nil), r.tasks...)
	r.mu.Unlock()
	return entries
}
