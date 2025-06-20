package main

import (
	"net/http/httptest"
	"testing"
)

func containsItem(items []IndexItem, name string) bool {
	for _, it := range items {
		if it.Name == name {
			return true
		}
	}
	return false
}

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := &CoreData{SecurityLevel: "administrator"}
	CustomNewsIndex(cd, req)
	if !containsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !containsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news")
	}

	cd = &CoreData{SecurityLevel: "writer"}
	CustomNewsIndex(cd, req)
	if containsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("writer should not see user permissions")
	}
	if !containsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("writer should see add news")
	}

	cd = &CoreData{SecurityLevel: "reader"}
	CustomNewsIndex(cd, req)
	if containsItem(cd.CustomIndexItems, "User Permissions") || containsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("reader should not see admin items")
	}
}

func TestCustomBlogIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs", nil)

	cd := &CoreData{SecurityLevel: "administrator"}
	CustomBlogIndex(cd, req)
	if !containsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !containsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("admin should see write blog")
	}

	cd = &CoreData{SecurityLevel: "writer"}
	CustomBlogIndex(cd, req)
	if containsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("writer should not see user permissions")
	}
	if !containsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("writer should see write blog")
	}

	cd = &CoreData{SecurityLevel: "reader"}
	CustomBlogIndex(cd, req)
	if containsItem(cd.CustomIndexItems, "User Permissions") || containsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("reader should not see writer/admin items")
	}
}
