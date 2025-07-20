package tasks

import "sync"

// NamedTask identifies a task by name.
type NamedTask interface {
	Name() string
}

var (
	regMu    sync.Mutex
	registry []NamedTask
)

// Register adds t to the registry. Duplicate names are ignored.
func Register(t NamedTask) {
	regMu.Lock()
	defer regMu.Unlock()
	for _, r := range registry {
		if r.Name() == t.Name() {
			return
		}
	}
	registry = append(registry, t)
}

// Registered returns a copy of the registered tasks slice.
func Registered() []NamedTask {
	regMu.Lock()
	tasks := append([]NamedTask(nil), registry...)
	regMu.Unlock()
	return tasks
}
