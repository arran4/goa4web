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
	if err == nil || !strings.Contains(err.Error(), "resolution is not supported") {
		t.Errorf("expected invalid parse for forum/topic ItemRef, got: %v", err)
	}
}

func TestForumLabelOperations(t *testing.T) {
	addOp := &ForumLabelAddOp{}
	removeOp := &ForumLabelRemoveOp{}

	// 1. Thread target with explicit ItemType
	for _, op := range []Operation{addOp, removeOp} {
		h := NewHeader()
		h.Set("Actor", "alice")
		h.Set("ItemType", "thread")
		h.Set("ItemRef", "staff-welcome")
		h.Set("Label", "team-priority")
		h.Set("At", "2026-08-01T10:00:00Z")
		evt := &Event{Headers: h}

		data, err := op.Parse(evt)
		if err != nil {
			t.Fatalf("%s: expected successful parse with ItemType: thread, got: %v", op.OpName(), err)
		}
		labelData, ok := data.(*forumLabelDataWithOp)
		if !ok || labelData.ItemType != "thread" || labelData.ItemRef != "staff-welcome" || labelData.Label != "team-priority" {
			t.Fatalf("%s: unexpected parsed data: %+v", op.OpName(), data)
		}

		refs := op.ReferencedSymbols(evt)
		if len(refs) != 2 {
			t.Fatalf("%s: expected 2 referenced symbols, got %d", op.OpName(), len(refs))
		}
		if refs[0].Type != RefTypeUser || refs[0].Symbol != "alice" {
			t.Errorf("%s: expected user alice, got %+v", op.OpName(), refs[0])
		}
		if refs[1].Type != RefTypeThread || refs[1].Symbol != "staff-welcome" {
			t.Errorf("%s: expected thread staff-welcome, got %+v", op.OpName(), refs[1])
		}
	}

	// 2. Topic target with explicit ItemType
	for _, op := range []Operation{addOp, removeOp} {
		h := NewHeader()
		h.Set("Actor", "alice")
		h.Set("ItemType", "topic")
		h.Set("ItemRef", "staff-room")
		h.Set("Label", "staff-priority")
		h.Set("At", "2026-08-01T10:00:00Z")
		evt := &Event{Headers: h}

		data, err := op.Parse(evt)
		if err != nil {
			t.Fatalf("%s: expected successful parse with ItemType: topic, got: %v", op.OpName(), err)
		}
		labelData, ok := data.(*forumLabelDataWithOp)
		if !ok || labelData.ItemType != "topic" || labelData.ItemRef != "staff-room" || labelData.Label != "staff-priority" {
			t.Fatalf("%s: unexpected parsed data: %+v", op.OpName(), data)
		}

		refs := op.ReferencedSymbols(evt)
		if len(refs) != 2 {
			t.Fatalf("%s: expected 2 referenced symbols, got %d", op.OpName(), len(refs))
		}
		if refs[0].Type != RefTypeUser || refs[0].Symbol != "alice" {
			t.Errorf("%s: expected user alice, got %+v", op.OpName(), refs[0])
		}
		if refs[1].Type != RefTypeForum || refs[1].Symbol != "staff-room" {
			t.Errorf("%s: expected forum topic staff-room, got %+v", op.OpName(), refs[1])
		}
	}

	// 3. Omitted ItemType defaults to thread for backward compatibility
	for _, op := range []Operation{addOp, removeOp} {
		h := NewHeader()
		h.Set("Actor", "alice")
		h.Set("ItemRef", "staff-welcome")
		h.Set("Label", "team-priority")
		h.Set("At", "2026-08-01T10:00:00Z")
		evt := &Event{Headers: h}

		data, err := op.Parse(evt)
		if err != nil {
			t.Fatalf("%s: expected successful parse with omitted ItemType, got: %v", op.OpName(), err)
		}
		labelData, ok := data.(*forumLabelDataWithOp)
		if !ok || labelData.ItemType != "thread" {
			t.Fatalf("%s: expected default ItemType thread, got %q", op.OpName(), labelData.ItemType)
		}

		refs := op.ReferencedSymbols(evt)
		if len(refs) != 2 || refs[1].Type != RefTypeThread {
			t.Errorf("%s: expected RefTypeThread for omitted ItemType, got %+v", op.OpName(), refs)
		}
	}

	// 4. Invalid ItemType rejected
	for _, op := range []Operation{addOp, removeOp} {
		h := NewHeader()
		h.Set("Actor", "alice")
		h.Set("ItemType", "invalid-type")
		h.Set("ItemRef", "foo")
		h.Set("Label", "bar")
		h.Set("At", "2026-08-01T10:00:00Z")
		evt := &Event{Headers: h}

		_, err := op.Parse(evt)
		if err == nil || !strings.Contains(err.Error(), "invalid ItemType") {
			t.Fatalf("%s: expected error containing 'invalid ItemType', got: %v", op.OpName(), err)
		}
	}
}
