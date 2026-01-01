package common_test

import (
	"reflect"
	"testing"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestSetThreadPublicLabels(t *testing.T) {
	q := &db.QuerierStub{
		ContentPublicLabels: map[string]map[int32]map[string]struct{}{
			"thread": {
				1: {"foo": {}, "bar": {}},
			},
		},
		ContentLabelStatus: map[string]map[int32]map[string]struct{}{
			"thread": {
				1: {},
			},
		},
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	if err := cd.SetThreadPublicLabels(1, []string{"bar", "baz"}); err != nil {
		t.Fatalf("SetThreadPublicLabels: %v", err)
	}

	labels := q.ContentPublicLabels["thread"][1]
	if len(labels) != 2 {
		t.Fatalf("labels %+v", labels)
	}
	if _, ok := labels["bar"]; !ok {
		t.Fatalf("missing bar label %+v", labels)
	}
	if _, ok := labels["baz"]; !ok {
		t.Fatalf("missing baz label %+v", labels)
	}
	if len(q.ListContentPublicLabelsCalls) != 1 {
		t.Fatalf("list public labels calls = %d", len(q.ListContentPublicLabelsCalls))
	}
	if len(q.ListContentLabelStatusCalls) != 1 {
		t.Fatalf("list author label calls = %d", len(q.ListContentLabelStatusCalls))
	}
	if len(q.AddContentPublicLabelCalls) != 1 || q.AddContentPublicLabelCalls[0].Label != "baz" {
		t.Fatalf("add public calls %+v", q.AddContentPublicLabelCalls)
	}
	if len(q.RemoveContentPublicLabelCalls) != 1 || q.RemoveContentPublicLabelCalls[0].Label != "foo" {
		t.Fatalf("remove public calls %+v", q.RemoveContentPublicLabelCalls)
	}
}

func TestSetThreadPrivateLabels(t *testing.T) {
	q := &db.QuerierStub{
		ContentPrivateLabels: map[string]map[int32]map[int32]map[string]bool{
			"thread": {
				1: {
					2: {"one": false, "two": false},
				},
			},
		},
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	if err := cd.SetThreadPrivateLabels(1, []string{"two", "three"}); err != nil {
		t.Fatalf("SetThreadPrivateLabels: %v", err)
	}

	labels := q.ContentPrivateLabels["thread"][1][2]
	twoInvert, ok := labels["two"]
	if len(labels) != 2 || !ok || twoInvert {
		t.Fatalf("private labels %+v", labels)
	}
	v, ok := labels["three"]
	if !ok || v {
		t.Fatalf("private labels %+v", labels)
	}
	if len(q.ListContentPrivateLabelsCalls) != 1 {
		t.Fatalf("list private labels calls = %d", len(q.ListContentPrivateLabelsCalls))
	}
	if len(q.AddContentPrivateLabelCalls) != 1 || q.AddContentPrivateLabelCalls[0].Label != "three" || q.AddContentPrivateLabelCalls[0].Invert {
		t.Fatalf("add private calls %+v", q.AddContentPrivateLabelCalls)
	}
	if len(q.RemoveContentPrivateLabelCalls) != 1 || q.RemoveContentPrivateLabelCalls[0].Label != "one" {
		t.Fatalf("remove private calls %+v", q.RemoveContentPrivateLabelCalls)
	}
}

func TestPrivateLabelsDefaultAndInversion(t *testing.T) {
	q := &db.QuerierStub{}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	t.Run("Default labels for thread", func(t *testing.T) {
		labels, err := cd.PrivateLabels("thread", 1, 1)
		if err != nil {
			t.Fatalf("PrivateLabels default: %v", err)
		}
		expected := []string{"new", "unread"}
		if !reflect.DeepEqual(labels, expected) {
			t.Fatalf("default labels %+v, want %+v", labels, expected)
		}
	})

	t.Run("Default labels for thread author", func(t *testing.T) {
		labels, err := cd.PrivateLabels("thread", 1, 2)
		if err != nil {
			t.Fatalf("PrivateLabels default: %v", err)
		}
		if len(labels) != 0 {
			t.Fatalf("default labels %+v, want empty", labels)
		}
	})

	if len(q.ListContentPrivateLabelsCalls) != 2 {
		t.Fatalf("list private labels calls = %d", len(q.ListContentPrivateLabelsCalls))
	}

	q.ContentPrivateLabels = map[string]map[int32]map[int32]map[string]bool{
		"thread": {
			1: {
				2: {"new": true, "foo": false},
			},
		},
	}
	q.ListContentPrivateLabelsCalls = nil

	labels, err := cd.PrivateLabels("thread", 1, 1)
	if err != nil {
		t.Fatalf("PrivateLabels invert: %v", err)
	}
	expected := []string{"unread", "foo"}
	if !reflect.DeepEqual(labels, expected) {
		t.Fatalf("inverted labels %+v, want %+v", labels, expected)
	}

	if len(q.ListContentPrivateLabelsCalls) != 1 {
		t.Fatalf("list private labels calls after reset = %d", len(q.ListContentPrivateLabelsCalls))
	}
}

func TestClearThreadPrivateLabelStatus(t *testing.T) {
	q := &db.QuerierStub{
		ContentPrivateLabels: map[string]map[int32]map[int32]map[string]bool{
			"thread": {
				1: {
					1: {"unread": true, "keep": false},
					2: {"unread": false},
				},
			},
		},
	}
	cd := common.NewTestCoreData(t, q)

	if err := cd.ClearThreadPrivateLabelStatus(1); err != nil {
		t.Fatalf("ClearThreadPrivateLabelStatus: %v", err)
	}

	if len(q.SystemClearContentPrivateLabelCalls) != 1 {
		t.Fatalf("system clear calls = %d", len(q.SystemClearContentPrivateLabelCalls))
	}
	remaining := q.ContentPrivateLabels["thread"][1]
	if _, ok := remaining[1]["unread"]; ok {
		t.Fatalf("expected unread cleared for user 1")
	}
	if _, ok := remaining[2]; ok {
		t.Fatalf("expected user 2 labels cleared entirely")
	}
	if v, ok := remaining[1]["keep"]; !ok || v {
		t.Fatalf("expected keep label retained for user 1")
	}
}

func TestPrivateLabelsTopicExcludesStatus(t *testing.T) {
	q := &db.QuerierStub{}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	labels, err := cd.PrivateLabels("topic", 1, 1)
	if err != nil {
		t.Fatalf("PrivateLabels topic: %v", err)
	}
	if len(labels) != 0 {
		t.Fatalf("expected no labels for topic, got %+v", labels)
	}

	if len(q.ListContentPrivateLabelsCalls) != 1 {
		t.Fatalf("list private labels calls = %d", len(q.ListContentPrivateLabelsCalls))
	}
}

func TestSetWritingPublicLabels(t *testing.T) {
	q := &db.QuerierStub{
		ContentLabelStatus: map[string]map[int32]map[string]struct{}{
			"writing": {
				5: {"a": {}, "b": {}},
			},
		},
		ContentPublicLabels: map[string]map[int32]map[string]struct{}{
			"writing": {
				5: {},
			},
		},
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	if err := cd.SetWritingAuthorLabels(5, []string{"b", "c"}); err != nil {
		t.Fatalf("SetWritingAuthorLabels: %v", err)
	}

	labels := q.ContentLabelStatus["writing"][5]
	if len(labels) != 2 {
		t.Fatalf("author labels %+v", labels)
	}
	if _, ok := labels["b"]; !ok {
		t.Fatalf("missing b label %+v", labels)
	}
	if _, ok := labels["c"]; !ok {
		t.Fatalf("missing c label %+v", labels)
	}
	if len(q.ListContentPublicLabelsCalls) != 1 {
		t.Fatalf("list public labels calls = %d", len(q.ListContentPublicLabelsCalls))
	}
	if len(q.ListContentLabelStatusCalls) != 1 {
		t.Fatalf("list author labels calls = %d", len(q.ListContentLabelStatusCalls))
	}
	if len(q.AddContentLabelStatusCalls) != 1 || q.AddContentLabelStatusCalls[0].Label != "c" {
		t.Fatalf("add author calls %+v", q.AddContentLabelStatusCalls)
	}
	if len(q.RemoveContentLabelStatusCalls) != 1 || q.RemoveContentLabelStatusCalls[0].Label != "a" {
		t.Fatalf("remove author calls %+v", q.RemoveContentLabelStatusCalls)
	}
}
