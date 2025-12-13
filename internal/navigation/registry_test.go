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

func TestIndexItemsPermissionFilter(t *testing.T) {
	defaultRegistry = NewRegistry()
	t.Cleanup(func() { defaultRegistry = NewRegistry() })

	RegisterIndexLinkWithViewPermission("protected", "/protected", 5, "news", "post")
	RegisterIndexLink("public", "/public", 10)

	items := IndexItemsWithPermission(func(section, item string) bool {
		return section == "news" && item == "post"
	})
	if len(items) != 2 {
		t.Fatalf("expected 2 items when permission is granted, got %d", len(items))
	}

	items = IndexItemsWithPermission(func(section, item string) bool {
		return false
	})
	if len(items) != 1 || items[0].Name != "public" {
		t.Fatalf("expected only public item when permission denied, got %#v", items)
	}
}
