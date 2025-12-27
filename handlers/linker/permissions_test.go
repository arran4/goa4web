package linker

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestUserCanCreateLink_Allowed(t *testing.T) {
	q := &db.QuerierStub{}

	ok, err := UserCanCreateLink(context.Background(), q, sql.NullInt32{Int32: 1, Valid: true}, 2)
	if err != nil {
		t.Fatalf("UserCanCreateLink: %v", err)
	}
	if !ok {
		t.Errorf("expected allowed")
	}
	if got := q.SystemCheckGrantCalls; len(got) != 1 {
		t.Fatalf("expected 1 SystemCheckGrant call, got %d", len(got))
	} else {
		call := got[0]
		if call.ViewerID != 2 {
			t.Errorf("unexpected viewer id %d", call.ViewerID)
		}
		if call.Section != "linker" || call.Action != "post" {
			t.Errorf("unexpected grant scope %q %q", call.Section, call.Action)
		}
		if call.Item.String != "category" || !call.Item.Valid {
			t.Errorf("unexpected item %v", call.Item)
		}
		if call.ItemID.Int32 != 1 || !call.ItemID.Valid {
			t.Errorf("unexpected item id %v", call.ItemID)
		}
		if call.UserID.Int32 != 2 || !call.UserID.Valid {
			t.Errorf("unexpected user id %v", call.UserID)
		}
	}
}

func TestUserCanCreateLink_Denied(t *testing.T) {
	q := &db.QuerierStub{SystemCheckGrantErr: sql.ErrNoRows}

	ok, err := UserCanCreateLink(context.Background(), q, sql.NullInt32{Int32: 1, Valid: true}, 2)
	if err != nil {
		t.Fatalf("UserCanCreateLink: %v", err)
	}
	if ok {
		t.Errorf("expected denied")
	}
	if got := q.SystemCheckGrantCalls; len(got) != 1 {
		t.Fatalf("expected 1 SystemCheckGrant call, got %d", len(got))
	}
}
