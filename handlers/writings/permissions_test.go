package writings

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/db/dbtest"
)

func TestUserCanCreateWriting(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{}

	ok, err := UserCanCreateWriting(ctx, q, 6, 18)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected permission granted")
	}
	requireWritingGrant(t, q, sql.NullInt32{Int32: 6, Valid: true}, 18)
}

func TestUserCanCreateWritingDenied(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{GrantResults: []error{sql.ErrNoRows}}

	ok, err := UserCanCreateWriting(ctx, q, 7, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected permission denied")
	}
	requireWritingGrant(t, q, sql.NullInt32{Int32: 7, Valid: true}, 1)
}

func TestUserCanCreateWritingError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("grant lookup failed")
	q := &dbtest.GrantLookupQuerier{GrantResults: []error{wantErr}}

	ok, err := UserCanCreateWriting(ctx, q, 9, 10)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v got %v", wantErr, err)
	}
	if ok {
		t.Fatalf("expected permission denied on error")
	}
	requireWritingGrant(t, q, sql.NullInt32{Int32: 9, Valid: true}, 10)
}

func requireWritingGrant(t *testing.T, q *dbtest.GrantLookupQuerier, itemID sql.NullInt32, viewerID int32) {
	t.Helper()
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one grant call, got %d", len(q.SystemCheckGrantCalls))
	}
	call := q.SystemCheckGrantCalls[0]
	if call.Section != "writing" || call.Item.String != "category" || call.Action != "post" {
		t.Fatalf("unexpected call %+v", call)
	}
	if call.ItemID != itemID {
		t.Fatalf("unexpected item id %v", call.ItemID)
	}
	wantUserID := sql.NullInt32{Int32: viewerID, Valid: viewerID != 0}
	if call.UserID != wantUserID || call.ViewerID != viewerID {
		t.Fatalf("unexpected viewer/user ids %+v", call)
	}
}
