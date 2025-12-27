package common_test

import (
	"reflect"
	"testing"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestSetThreadPublicLabels(t *testing.T) {
	q := &db.QuerierStub{
		ContentPublicLabelsRows: map[string][]*db.ListContentPublicLabelsRow{
			"thread:1": {
				{Item: "thread", ItemID: 1, Label: "foo"},
				{Item: "thread", ItemID: 1, Label: "bar"},
			},
		},
		ContentLabelStatusRows: map[string][]*db.ListContentLabelStatusRow{
			"thread:1": {},
		},
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	if err := cd.SetThreadPublicLabels(1, []string{"bar", "baz"}); err != nil {
		t.Fatalf("SetThreadPublicLabels: %v", err)
	}

	if got := len(q.ListContentPublicLabelsCalls); got != 1 {
		t.Fatalf("public label lookup calls %d, want 1", got)
	}
	if got := q.AddContentPublicLabelCalls; len(got) != 1 || got[0].Label != "baz" {
		t.Fatalf("public label inserts %+v, want [baz]", got)
	}
	if got := q.RemoveContentPublicLabelCalls; len(got) != 1 || got[0].Label != "foo" {
		t.Fatalf("public label removals %+v, want [foo]", got)
	}
}

func TestSetThreadPrivateLabels(t *testing.T) {
	q := &db.QuerierStub{
		ContentPrivateLabelsRows: map[string][]*db.ListContentPrivateLabelsRow{
			"thread:1:2": {
				{Item: "thread", ItemID: 1, UserID: 2, Label: "one"},
				{Item: "thread", ItemID: 1, UserID: 2, Label: "two"},
			},
		},
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	if err := cd.SetThreadPrivateLabels(1, []string{"two", "three"}); err != nil {
		t.Fatalf("SetThreadPrivateLabels: %v", err)
	}

	if got := q.AddContentPrivateLabelCalls; len(got) != 1 || got[0].Label != "three" || got[0].Invert {
		t.Fatalf("private label inserts %+v, want [three false]", got)
	}
	if got := q.RemoveContentPrivateLabelCalls; len(got) != 1 || got[0].Label != "one" {
		t.Fatalf("private label removals %+v, want [one]", got)
	}
}

func TestPrivateLabelsDefaultAndInversion(t *testing.T) {
	q := &db.QuerierStub{
		ContentPrivateLabelsRows: map[string][]*db.ListContentPrivateLabelsRow{
			"thread:1:2": {},
		},
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	labels, err := cd.PrivateLabels("thread", 1)
	if err != nil {
		t.Fatalf("PrivateLabels default: %v", err)
	}
	expected := []string{"new", "unread"}
	if !reflect.DeepEqual(labels, expected) {
		t.Fatalf("default labels %+v, want %+v", labels, expected)
	}

	// Inversion case: storing an inverted new label removes it from the result.
	q.ContentPrivateLabelsRows["thread:1:2"] = []*db.ListContentPrivateLabelsRow{
		{Item: "thread", ItemID: 1, UserID: 2, Label: "new", Invert: true},
		{Item: "thread", ItemID: 1, UserID: 2, Label: "foo"},
	}

	labels, err = cd.PrivateLabels("thread", 1)
	if err != nil {
		t.Fatalf("PrivateLabels invert: %v", err)
	}
	expected = []string{"unread", "foo"}
	if !reflect.DeepEqual(labels, expected) {
		t.Fatalf("inverted labels %+v, want %+v", labels, expected)
	}
}

func TestClearThreadPrivateLabelStatus(t *testing.T) {
	q := &db.QuerierStub{}
	cd := common.NewTestCoreData(t, q)

	if err := cd.ClearThreadPrivateLabelStatus(1); err != nil {
		t.Fatalf("ClearThreadPrivateLabelStatus: %v", err)
	}

	if got := q.SystemClearContentPrivateLabelCalls; len(got) != 1 {
		t.Fatalf("clear label status calls %+v, want 1 call", got)
	} else if got[0].Label != "unread" || got[0].Item != "thread" || got[0].ItemID != 1 {
		t.Fatalf("clear label status args %+v, want thread:1 unread", got[0])
	}
}

func TestPrivateLabelsTopicExcludesStatus(t *testing.T) {
	q := &db.QuerierStub{
		ContentPrivateLabelsRows: map[string][]*db.ListContentPrivateLabelsRow{
			"topic:1:2": {},
		},
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	labels, err := cd.PrivateLabels("topic", 1)
	if err != nil {
		t.Fatalf("PrivateLabels topic: %v", err)
	}
	if len(labels) != 0 {
		t.Fatalf("expected no labels for topic, got %+v", labels)
	}
}

func TestSetWritingPublicLabels(t *testing.T) {
	q := &db.QuerierStub{
		ContentPublicLabelsRows: map[string][]*db.ListContentPublicLabelsRow{
			"writing:5": {},
		},
		ContentLabelStatusRows: map[string][]*db.ListContentLabelStatusRow{
			"writing:5": {
				{Item: "writing", ItemID: 5, Label: "a"},
				{Item: "writing", ItemID: 5, Label: "b"},
			},
		},
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	if err := cd.SetWritingAuthorLabels(5, []string{"b", "c"}); err != nil {
		t.Fatalf("SetWritingAuthorLabels: %v", err)
	}

	if got := q.AddContentLabelStatusCalls; len(got) != 1 || got[0].Label != "c" {
		t.Fatalf("author label inserts %+v, want [c]", got)
	}
	if got := q.RemoveContentLabelStatusCalls; len(got) != 1 || got[0].Label != "a" {
		t.Fatalf("author label removals %+v, want [a]", got)
	}
}
