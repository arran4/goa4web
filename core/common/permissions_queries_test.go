package common_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db/dbtest"
)

func TestHasGrantQueriesDatabase(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{}
	cd := common.NewCoreData(ctx, q, nil)
	cd.UserID = 7

	if !cd.HasGrant("forum", "topic", "reply", 13) {
		t.Fatalf("expected HasGrant to return true when the grant check succeeds")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one SystemCheckGrant call, got %d", len(q.SystemCheckGrantCalls))
	}
	call := q.SystemCheckGrantCalls[0]
	if call.ViewerID != 7 || call.Section != "forum" || call.Action != "reply" {
		t.Fatalf("unexpected call %#v", call)
	}
	if call.Item.String != "topic" || !call.Item.Valid {
		t.Fatalf("expected topic item in call, got %#v", call.Item)
	}
	wantItemID := sql.NullInt32{Int32: 13, Valid: true}
	if call.ItemID != wantItemID {
		t.Fatalf("expected item id %v, got %v", wantItemID, call.ItemID)
	}
	wantUserID := sql.NullInt32{Int32: 7, Valid: true}
	if call.UserID != wantUserID {
		t.Fatalf("expected user id %v, got %v", wantUserID, call.UserID)
	}
}

func TestHasGrantReturnsFalseOnGrantError(t *testing.T) {
	ctx := context.Background()
	q := &dbtest.GrantLookupQuerier{GrantResults: []error{sql.ErrNoRows}}
	cd := common.NewCoreData(ctx, q, nil)
	cd.UserID = 3

	if cd.HasGrant("forum", "topic", "post", 99) {
		t.Fatalf("expected HasGrant to return false on grant lookup error")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one SystemCheckGrant call, got %d", len(q.SystemCheckGrantCalls))
	}
	call := q.SystemCheckGrantCalls[0]
	if call.ViewerID != 3 {
		t.Fatalf("expected viewer id 3, got %d", call.ViewerID)
	}
}
