package scenario

import (
	"strings"
	"testing"
)

func TestUserGrantReferencedSymbols(t *testing.T) {
	op := &UserGrantOp{}

	// Test a valid thread reference
	h := NewHeader()
	h.Set("User", "alice")
	h.Set("Section", "privateforum_thread")
	h.Set("Item", "thread")
	h.Set("ItemRef", "valid-ref")
	h.Set("Action", "append")
	evt := &Event{Headers: h}

	syms := op.ReferencedSymbols(evt)
	if len(syms) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(syms))
	}
	if syms[1].Symbol != "valid-ref" {
		t.Errorf("expected valid-ref, got %s", syms[1].Symbol)
	}
}

func TestUserGrantApplyLogicItemRef(t *testing.T) {
	op := &UserGrantOp{}
	h := NewHeader()
	h.Set("User", "alice")
	h.Set("Section", "privateforum_thread")
	h.Set("Item", "thread")
	h.Set("ItemRef", "staff-welcome")
	h.Set("Action", "append")
	evt := &Event{Headers: h}

	// Valid tuple
	_, err := op.Parse(evt)
	if err != nil {
		t.Errorf("expected valid Parse, got %v", err)
	}

	// Incompatible reference missing
	h.Set("ItemRef", "")
	evt.Headers = h
	_, err = op.Parse(evt)
	if err == nil || !strings.Contains(err.Error(), "ItemRef is required") {
		t.Errorf("expected error for missing item ref, got %v", err)
	}

	// Incompatible reference included for global action
	h.Set("Section", "privateforum")
	h.Set("Item", "topic")
	h.Set("ItemRef", "wrong-ref")
	h.Set("Action", "view")
	evt.Headers = h
	_, err = op.Parse(evt)
	if err == nil || !strings.Contains(err.Error(), "does not support or require an item ID") {
		t.Errorf("expected error for global action with ItemRef, got %v", err)
	}
}
func TestUserGrantParse(t *testing.T) {
	op := &UserGrantOp{}
	h := NewHeader()
	h.Set("User", "alice")
	h.Set("Section", "privateforum_thread")
	h.Set("Item", "thread")
	h.Set("ItemRef", "missing-ref")
	h.Set("Action", "append")
	h.Set("At", "2026-08-01T09:16:00+10:00")
	evt := &Event{Headers: h}

	_, err := op.Parse(evt)
	if err != nil {
		t.Errorf("expected nil error on valid parse, got %v", err)
	}

	h.Set("Section", "invalid_section")
	evt.Headers = h
	_, err = op.Parse(evt)
	if err == nil || !strings.Contains(err.Error(), "unsupported permission tuple") {
		t.Errorf("expected error on unsupported permission tuple, got %v", err)
	}

	h.Set("Section", "privateforum_thread")
	h.Set("Item", "thread")
	h.Set("ItemRef", "") // missing item ref where required
	evt.Headers = h
	_, err = op.Parse(evt)
	if err == nil || !strings.Contains(err.Error(), "ItemRef is required") {
		t.Errorf("expected error for missing ItemRef on item-specific permission, got: %v", err)
	}

	h.Set("Section", "forum")
	h.Set("Item", "topic")
	h.Set("ItemRef", "test-topic-ref")
	evt.Headers = h
	_, err = op.Parse(evt)
	if err != nil {
		t.Errorf("expected valid parse for forum/topic ItemRef, got: %v", err)
	}
}
