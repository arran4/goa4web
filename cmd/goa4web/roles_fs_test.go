package main

import (
	"testing"

	"github.com/arran4/goa4web/internal/roles"
)

func TestListAndReadEmbeddedRoles(t *testing.T) {
	rolesList, err := roles.ListEmbeddedRoles()
	if err != nil {
		t.Fatalf("listEmbeddedRoles error: %v", err)
	}
	if len(rolesList) == 0 {
		t.Fatalf("expected at least one embedded role, got 0")
	}

	// Try reading the first role
	if _, err := roles.ReadEmbeddedRole(rolesList[0]); err != nil {
		t.Fatalf("readEmbeddedRole(%q) error: %v", rolesList[0], err)
	}

	// Unknown role should error
	if _, err := roles.ReadEmbeddedRole("__nope__"); err == nil {
		t.Fatalf("expected error for unknown role, got nil")
	}
}
