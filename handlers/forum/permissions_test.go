package forum

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/db/dbtest"
)

func TestUserCanCreateThread(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{}

	ok, err := UserCanCreateThread(ctx, q, "forum", 5, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected permission granted")
	}
	requireGrantCall(t, q, "forum", "topic", "post", sql.NullInt32{Int32: 5, Valid: true}, 42)
}

func TestUserCanCreateThreadDenied(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{GrantResults: []error{sql.ErrNoRows}}

	ok, err := UserCanCreateThread(ctx, q, "forum", 7, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected permission denied")
	}
	requireGrantCall(t, q, "forum", "topic", "post", sql.NullInt32{Int32: 7, Valid: true}, 9)
}

func TestUserCanCreateThreadError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("boom")
	q := &dbtest.GrantLookupQuerier{GrantResults: []error{wantErr}}

	ok, err := UserCanCreateThread(ctx, q, "forum", 3, 11)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v got %v", wantErr, err)
	}
	if ok {
		t.Fatalf("expected permission denied on error")
	}
	requireGrantCall(t, q, "forum", "topic", "post", sql.NullInt32{Int32: 3, Valid: true}, 11)
}

func TestUserCanCreateTopic(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{}

	ok, err := UserCanCreateTopic(ctx, q, "forum", 8, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected permission granted")
	}
	requireGrantCall(t, q, "forum", "category", "post", sql.NullInt32{Int32: 8, Valid: true}, 4)
}

func TestUserCanCreateTopicDenied(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{GrantResults: []error{sql.ErrNoRows}}

	ok, err := UserCanCreateTopic(ctx, q, "forum", 2, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected permission denied")
	}
	requireGrantCall(t, q, "forum", "category", "post", sql.NullInt32{Int32: 2, Valid: true}, 6)
}

func TestUserCanCreateTopicError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("grant failure")
	q := &dbtest.GrantLookupQuerier{GrantResults: []error{wantErr}}

	ok, err := UserCanCreateTopic(ctx, q, "forum", 10, 14)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v got %v", wantErr, err)
	}
	if ok {
		t.Fatalf("expected permission denied on error")
	}
	requireGrantCall(t, q, "forum", "category", "post", sql.NullInt32{Int32: 10, Valid: true}, 14)
}

func requireGrantCall(t *testing.T, q *dbtest.GrantLookupQuerier, section, item, action string, itemID sql.NullInt32, viewerID int32) {
	t.Helper()
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one grant lookup, got %d", len(q.SystemCheckGrantCalls))
	}
	call := q.SystemCheckGrantCalls[0]
	if call.Section != section || call.Action != action || call.ViewerID != viewerID {
		t.Fatalf("unexpected call %+v", call)
	}
	if call.Item.String != item || !call.Item.Valid {
		t.Fatalf("unexpected item %+v", call.Item)
	}
	if call.ItemID != itemID {
		t.Fatalf("unexpected itemID %v", call.ItemID)
	}
	wantUserID := sql.NullInt32{Int32: viewerID, Valid: viewerID != 0}
	if call.UserID != wantUserID {
		t.Fatalf("unexpected userID %v", call.UserID)
	}
}
