package common_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/arran4/goa4web/config"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/db/testutil"
)

func TestSetThreadPublicLabels(t *testing.T) {
	queries := testutil.NewLabelsQuerier(t)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 2

	queries.PublicLabels = []*db.ListContentPublicLabelsRow{
		{Item: "thread", ItemID: 1, Label: "foo"},
		{Item: "thread", ItemID: 1, Label: "bar"},
	}
	queries.LabelStatus = nil

	if err := cd.SetThreadPublicLabels(1, []string{"bar", "baz"}); err != nil {
		t.Fatalf("SetThreadPublicLabels: %v", err)
	}
	if !reflect.DeepEqual(queries.AddedPublic, []string{"baz"}) {
		t.Fatalf("added labels %+v, want [baz]", queries.AddedPublic)
	}
	if !reflect.DeepEqual(queries.RemovedPublic, []string{"foo"}) {
		t.Fatalf("removed labels %+v, want [foo]", queries.RemovedPublic)
	}
}

func TestSetThreadPrivateLabels(t *testing.T) {
	queries := testutil.NewLabelsQuerier(t)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 2

	queries.PrivateLabels = []*db.ListContentPrivateLabelsRow{
		{Item: "thread", ItemID: 1, UserID: 2, Label: "one"},
		{Item: "thread", ItemID: 1, UserID: 2, Label: "two"},
	}

	if err := cd.SetThreadPrivateLabels(1, []string{"two", "three"}); err != nil {
		t.Fatalf("SetThreadPrivateLabels: %v", err)
	}
	if !reflect.DeepEqual(queries.AddedPrivate, []string{"three"}) {
		t.Fatalf("added labels %+v, want [three]", queries.AddedPrivate)
	}
	if !reflect.DeepEqual(queries.RemovedPrivate, []string{"one"}) {
		t.Fatalf("removed labels %+v, want [one]", queries.RemovedPrivate)
	}
}

func TestPrivateLabelsDefaultAndInversion(t *testing.T) {
	queries := testutil.NewLabelsQuerier(t)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 2

	// Default case: no stored rows should return new and unread labels.
	queries.PrivateLabels = nil

	labels, err := cd.PrivateLabels("thread", 1)
	if err != nil {
		t.Fatalf("PrivateLabels default: %v", err)
	}
	expected := []string{"new", "unread"}
	if !reflect.DeepEqual(labels, expected) {
		t.Fatalf("default labels %+v, want %+v", labels, expected)
	}

	// Inversion case: storing an inverted new label removes it from the result.
	queries.PrivateLabels = []*db.ListContentPrivateLabelsRow{
		{Item: "thread", ItemID: 1, UserID: 2, Label: "new", Invert: true},
		{Item: "thread", ItemID: 1, UserID: 2, Label: "foo", Invert: false},
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
	queries := testutil.NewLabelsQuerier(t)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	if err := cd.ClearThreadPrivateLabelStatus(1); err != nil {
		t.Fatalf("ClearThreadPrivateLabelStatus: %v", err)
	}
	if len(queries.RemovedPrivate) != 1 || queries.RemovedPrivate[0] != "unread" {
		t.Fatalf("removed labels %+v, want [unread]", queries.RemovedPrivate)
	}
}

func TestPrivateLabelsTopicExcludesStatus(t *testing.T) {
	queries := testutil.NewLabelsQuerier(t)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 2

	queries.PrivateLabels = nil

	labels, err := cd.PrivateLabels("topic", 1)
	if err != nil {
		t.Fatalf("PrivateLabels topic: %v", err)
	}
	if len(labels) != 0 {
		t.Fatalf("expected no labels for topic, got %+v", labels)
	}
}

func TestSetWritingPublicLabels(t *testing.T) {
	queries := testutil.NewLabelsQuerier(t)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 2

	queries.PublicLabels = nil
	queries.LabelStatus = []*db.ListContentLabelStatusRow{
		{Item: "writing", ItemID: 5, Label: "a"},
		{Item: "writing", ItemID: 5, Label: "b"},
	}

	if err := cd.SetWritingAuthorLabels(5, []string{"b", "c"}); err != nil {
		t.Fatalf("SetWritingAuthorLabels: %v", err)
	}
	if !reflect.DeepEqual(queries.AddedStatus, []string{"c"}) {
		t.Fatalf("added labels %+v, want [c]", queries.AddedStatus)
	}
	if !reflect.DeepEqual(queries.RemovedStatus, []string{"a"}) {
		t.Fatalf("removed labels %+v, want [a]", queries.RemovedStatus)
	}
}
