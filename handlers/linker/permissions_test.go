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
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one grant check, got %d", len(q.SystemCheckGrantCalls))
	}
	call := q.SystemCheckGrantCalls[0]
	if call.Action != "post" || call.Section != "linker" {
		t.Fatalf("unexpected grant params %+v", call)
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
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}
