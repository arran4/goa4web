package writings

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestUserCanCreateWriting_Allowed(t *testing.T) {
	q := &db.QuerierStub{}

	ok, err := UserCanCreateWriting(context.Background(), q, 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateWriting: %v", err)
	}
	if !ok {
		t.Errorf("expected allowed")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestUserCanCreateWriting_Denied(t *testing.T) {
	q := &db.QuerierStub{SystemCheckGrantErr: sql.ErrNoRows}

	ok, err := UserCanCreateWriting(context.Background(), q, 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateWriting: %v", err)
	}
	if ok {
		t.Errorf("expected denied")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestUserCanCreateWriting_Error(t *testing.T) {
	q := &db.QuerierStub{SystemCheckGrantErr: errors.New("db offline")}

	ok, err := UserCanCreateWriting(context.Background(), q, 1, 2)
	if err == nil {
		t.Fatalf("expected error")
	}
	if ok {
		t.Fatalf("expected denied when error occurs")
	}
}
