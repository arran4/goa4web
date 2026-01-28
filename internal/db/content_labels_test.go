package db_test

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestAddAndListContentPublicLabels(t *testing.T) {
	q := testhelpers.NewQuerierStub()

	if err := q.AddContentPublicLabel(context.Background(), db.AddContentPublicLabelParams{
		Item:   "thread",
		ItemID: 1,
		Label:  "foo",
	}); err != nil {
		t.Fatalf("AddContentPublicLabel: %v", err)
	}

	res, err := q.ListContentPublicLabels(context.Background(), db.ListContentPublicLabelsParams{Item: "thread", ItemID: 1})
	if err != nil {
		t.Fatalf("ListContentPublicLabels: %v", err)
	}
	if len(res) != 1 || res[0].Label != "foo" {
		t.Fatalf("unexpected result %+v", res)
	}
}

func TestAddAndListContentPrivateLabels(t *testing.T) {
	q := testhelpers.NewQuerierStub()

	if err := q.AddContentPrivateLabel(context.Background(), db.AddContentPrivateLabelParams{
		Item:   "thread",
		ItemID: 1,
		UserID: 2,
		Label:  "bar",
		Invert: false,
	}); err != nil {
		t.Fatalf("AddContentPrivateLabel: %v", err)
	}

	res, err := q.ListContentPrivateLabels(context.Background(), db.ListContentPrivateLabelsParams{Item: "thread", ItemID: 1, UserID: 2})
	if err != nil {
		t.Fatalf("ListContentPrivateLabels: %v", err)
	}
	if len(res) != 1 || res[0].Label != "bar" || res[0].Invert {
		t.Fatalf("unexpected result %+v", res)
	}
}
