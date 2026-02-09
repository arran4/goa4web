package tasks

import (
	"fmt"
	"sync"
	"testing"
)

type mockTask struct {
	name string
}

func (m mockTask) Name() string { return m.name }

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	task := mockTask{name: "Task1"}

	r.Register("SectionA", task)

	entries := r.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Section != "SectionA" {
		t.Errorf("expected section 'SectionA', got '%s'", entries[0].Section)
	}
	if entries[0].Task.Name() != "Task1" {
		t.Errorf("expected task name 'Task1', got '%s'", entries[0].Task.Name())
	}

	registered := r.Registered()
	if len(registered) != 1 {
		t.Fatalf("expected 1 registered task, got %d", len(registered))
	}
	if registered[0].Name() != "Task1" {
		t.Errorf("expected registered task name 'Task1', got '%s'", registered[0].Name())
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	task := mockTask{name: "Task1"}

	r.Register("SectionA", task)
	r.Register("SectionA", task) // Duplicate

	entries := r.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestRegistry_RegisterDifferentSections(t *testing.T) {
	r := NewRegistry()
	task := mockTask{name: "Task1"}

	r.Register("SectionA", task)
	r.Register("SectionB", task)

	entries := r.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestRegistry_Concurrency(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	count := 100

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			task := mockTask{name: fmt.Sprintf("Task%d", i)}
			r.Register("SectionA", task)
		}(i)
	}

	wg.Wait()

	entries := r.Entries()
	if len(entries) != count {
		t.Errorf("expected %d entries, got %d", count, len(entries))
	}
}
