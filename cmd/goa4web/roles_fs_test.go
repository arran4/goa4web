package main

import "testing"

func TestListAndReadEmbeddedRoles(t *testing.T) {
	roles, err := listEmbeddedRoles()
	if err != nil {
		t.Fatalf("listEmbeddedRoles error: %v", err)
	}
	if len(roles) == 0 {
		t.Fatalf("expected at least one embedded role, got 0")
	}

	// Try reading the first role
	if _, err := readEmbeddedRole(roles[0]); err != nil {
		t.Fatalf("readEmbeddedRole(%q) error: %v", roles[0], err)
	}

	// Unknown role should error
	if _, err := readEmbeddedRole("__nope__"); err == nil {
		t.Fatalf("expected error for unknown role, got nil")
	}
}
