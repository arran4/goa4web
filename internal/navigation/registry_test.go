package navigation

import (
	"testing"
)

func TestIndexItemsOrdering(t *testing.T) {
	defaultRegistry = NewRegistry()
	t.Cleanup(func() { defaultRegistry = NewRegistry() })

	RegisterIndexLink("b", "/b", 20)
	RegisterIndexLink("a", "/a", 10)
	RegisterAdminControlCenter("b", "/admin/b", 20)
	RegisterAdminControlCenter("a", "/admin/a", 10)

	items := IndexItems()
	if len(items) != 2 {
		t.Fatalf("want 2 items got %d", len(items))
	}
	if items[0].Name != "a" {
		t.Errorf("index 0 want a got %s", items[0].Name)
	}

	links := AdminLinks()
	if len(links) != 2 {
		t.Fatalf("want 2 links got %d", len(links))
	}
	if links[0].Name != "a" {
		t.Errorf("admin link 0 want a got %s", links[0].Name)
	}
}

func TestIndexItemsSkipEmpty(t *testing.T) {
	defaultRegistry = NewRegistry()
	t.Cleanup(func() { defaultRegistry = NewRegistry() })

	RegisterAdminControlCenter("no", "/admin/no", 5)
	items := IndexItems()
	if len(items) != 0 {
		t.Fatalf("expected 0 items got %d", len(items))
	}
	links := AdminLinks()
	if len(links) != 1 || links[0].Name != "no" {
		t.Fatalf("unexpected admin links %#v", links)
	}
}
