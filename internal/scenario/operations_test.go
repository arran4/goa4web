package scenario

import (
	"strings"
	"testing"
)

func TestUserGrantReferencedSymbols(t *testing.T) {
	op := &UserGrantOp{}
	h := NewHeader()
	h.Set("User", "alice")
	h.Set("Section", "privateforum_thread")
	h.Set("Item", "thread")
	h.Set("ItemRef", "missing-ref")
	h.Set("Action", "append")
	evt := &Event{Headers: h}

	syms := op.ReferencedSymbols(evt)
	if len(syms) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(syms))
	}

	if syms[1].Symbol != "missing-ref" {
		t.Errorf("expected missing-ref, got %s", syms[1].Symbol)
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
}
