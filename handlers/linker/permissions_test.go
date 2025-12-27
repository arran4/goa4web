package linker

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/db/dbtest"
)

func TestUserCanCreateLink(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{}

	ok, err := UserCanCreateLink(ctx, q, sql.NullInt32{Int32: 12, Valid: true}, 21)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected permission granted")
	}
	requireSingleGrant(t, q, "linker", "category", "post", sql.NullInt32{Int32: 12, Valid: true}, 21)
}

func TestUserCanCreateLinkDenied(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{GrantResults: []error{sql.ErrNoRows}}

	ok, err := UserCanCreateLink(ctx, q, sql.NullInt32{Int32: 2, Valid: true}, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected permission denied")
	}
	requireSingleGrant(t, q, "linker", "category", "post", sql.NullInt32{Int32: 2, Valid: true}, 3)
}

func TestUserCanCreateLinkError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("grant check failed")
	q := &dbtest.GrantLookupQuerier{GrantResults: []error{wantErr}}

	ok, err := UserCanCreateLink(ctx, q, sql.NullInt32{Int32: 4, Valid: true}, 5)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v got %v", wantErr, err)
	}
	if ok {
		t.Fatalf("expected permission denied on error")
	}
	requireSingleGrant(t, q, "linker", "category", "post", sql.NullInt32{Int32: 4, Valid: true}, 5)
}

func requireSingleGrant(t *testing.T, q *dbtest.GrantLookupQuerier, section, item, action string, itemID sql.NullInt32, viewerID int32) {
	t.Helper()
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one SystemCheckGrant call, got %d", len(q.SystemCheckGrantCalls))
	}
	call := q.SystemCheckGrantCalls[0]
	if call.Section != section || call.Action != action || call.ViewerID != viewerID {
		t.Fatalf("unexpected call %+v", call)
	}
	if call.Item.String != item || !call.Item.Valid {
		t.Fatalf("unexpected item %+v", call.Item)
	}
	if call.ItemID != itemID {
		t.Fatalf("unexpected item id %v", call.ItemID)
	}
	wantUserID := sql.NullInt32{Int32: viewerID, Valid: viewerID != 0}
	if call.UserID != wantUserID {
		t.Fatalf("unexpected user id %v", call.UserID)
	}
}
