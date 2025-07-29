package navigation

import (
	"testing"
)

func TestIndexItemsOrdering(t *testing.T) {
	defaultRegistry = NewRegistry()
	t.Cleanup(func() { defaultRegistry = NewRegistry() })

	RegisterIndexLink("b", "/b", 20)
	RegisterIndexLink("a", "/a", 10)
	RegisterAdminControlCenter("sec", "b", "/admin/b", 20)
	RegisterAdminControlCenter("sec", "a", "/admin/a", 10)

	items := IndexItems()
	if len(items) != 2 {
		t.Fatalf("want 2 items got %d", len(items))
	}
	if items[0].Name != "a" {
		t.Errorf("index 0 want a got %s", items[0].Name)
	}

	secs := AdminSections()
	if len(secs) != 1 {
		t.Fatalf("want 1 section got %d", len(secs))
	}
	if len(secs[0].Links) != 2 || secs[0].Links[0].Name != "a" {
		t.Fatalf("unexpected admin section %#v", secs)
	}
}

func TestIndexItemsSkipEmpty(t *testing.T) {
	defaultRegistry = NewRegistry()
	t.Cleanup(func() { defaultRegistry = NewRegistry() })

	RegisterAdminControlCenter("sec", "no", "/admin/no", 5)
	items := IndexItems()
	if len(items) != 0 {
		t.Fatalf("expected 0 items got %d", len(items))
	}
	secs := AdminSections()
	if len(secs) != 1 || len(secs[0].Links) != 1 || secs[0].Links[0].Name != "no" {
		t.Fatalf("unexpected admin sections %#v", secs)
	}
}
