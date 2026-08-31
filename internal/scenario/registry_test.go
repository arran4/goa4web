package scenario

import (
	"errors"
	"testing"
)

func TestRefRegistry(t *testing.T) {
	reg := NewRefRegistry()

	// Special built-in actors
	if !reg.HasDeclared(RefTypeUser, "admin") {
		t.Error("expected admin to be declared implicitly")
	}
	if !reg.HasDeclared(RefTypeUser, "system") {
		t.Error("expected system to be declared implicitly")
	}

	// Declare new user
	if err := reg.Declare(RefTypeUser, "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reg.HasDeclared(RefTypeUser, "alice") {
		t.Error("expected alice to be declared")
	}

	// Same name, different type is allowed (typed symbolic references)
	if err := reg.Declare(RefTypeForum, "alice"); err != nil {
		t.Fatalf("unexpected error declaring alice as forum: %v", err)
	}
	if !reg.HasDeclared(RefTypeForum, "alice") {
		t.Error("expected alice to be declared as forum")
	}

	// Duplicate declaration of same type rejected
	err := reg.Declare(RefTypeUser, "alice")
	if err == nil {
		t.Fatal("expected duplicate error, got nil")
	}
	var dup ErrDuplicateRef
	if !errors.As(err, &dup) || dup.Symbol != "alice" || dup.Type != RefTypeUser {
		t.Fatalf("expected ErrDuplicateRef for user alice, got %v", err)
	}

	// Bind and Resolve
	if err := reg.Bind(RefTypeUser, "alice", int32(42)); err != nil {
		t.Fatalf("bind error: %v", err)
	}
	val, ok := reg.Resolve(RefTypeUser, "alice")
	if !ok || val != int32(42) {
		t.Fatalf("expected 42, got %v (ok=%v)", val, ok)
	}

	uid, ok := reg.ResolveUser("alice")
	if !ok || uid != 42 {
		t.Fatalf("ResolveUser expected 42, got %d (ok=%v)", uid, ok)
	}

	// Unbound symbol
	if _, ok := reg.Resolve(RefTypeUser, "bob"); ok {
		t.Error("expected bob to be unresolvable")
	}
}
