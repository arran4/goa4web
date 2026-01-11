package forum

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestUserCanCreateThread_Allowed(t *testing.T) {
	q := &db.QuerierStub{SystemCheckGrantReturns: 1}

	ok, err := UserCanCreateThread(context.Background(), q, "forum", 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateThread: %v", err)
	}
	if !ok {
		t.Errorf("expected allowed")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
	call := q.SystemCheckGrantCalls[0]
	if call.Section != "forum" || call.Action != "post" {
		t.Fatalf("unexpected grant check params: %#v", call)
	}
	if !call.ItemID.Valid || call.ItemID.Int32 != 1 {
		t.Fatalf("unexpected topic id: %#v", call.ItemID)
	}
	if call.ViewerID != 2 || !call.UserID.Valid || call.UserID.Int32 != 2 {
		t.Fatalf("unexpected viewer/user: viewer=%d user=%#v", call.ViewerID, call.UserID)
	}
}

func TestUserCanCreateThread_Denied(t *testing.T) {
	q := &db.QuerierStub{SystemCheckGrantErr: sql.ErrNoRows}

	ok, err := UserCanCreateThread(context.Background(), q, "forum", 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateThread: %v", err)
	}
	if ok {
		t.Errorf("expected denied")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestUserCanCreateThread_Error(t *testing.T) {
	q := &db.QuerierStub{SystemCheckGrantErr: errors.New("db offline")}

	ok, err := UserCanCreateThread(context.Background(), q, "forum", 1, 2)
	if err == nil {
		t.Fatalf("expected error")
	}
	if ok {
		t.Fatalf("expected denied when error occurs")
	}
}

func TestUserCanCreateTopic_Allowed(t *testing.T) {
	q := &db.QuerierStub{SystemCheckGrantReturns: 1}

	ok, err := UserCanCreateTopic(context.Background(), q, "forum", 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateTopic: %v", err)
	}
	if !ok {
		t.Errorf("expected allowed")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestUserCanCreateTopic_Denied(t *testing.T) {
	q := &db.QuerierStub{SystemCheckGrantErr: sql.ErrNoRows}

	ok, err := UserCanCreateTopic(context.Background(), q, "forum", 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateTopic: %v", err)
	}
	if ok {
		t.Errorf("expected denied")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestUserCanCreateTopic_Error(t *testing.T) {
	q := &db.QuerierStub{SystemCheckGrantErr: errors.New("db offline")}

	ok, err := UserCanCreateTopic(context.Background(), q, "forum", 1, 2)
	if err == nil {
		t.Fatalf("expected error")
	}
	if ok {
		t.Fatalf("expected denied when error occurs")
	}
}
