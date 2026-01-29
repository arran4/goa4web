package forum

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestThreadDelete(t *testing.T) {
	q := testhelpers.NewQuerierStub()

	if err := ThreadDelete(context.Background(), q, 1, 2); err != nil {
		t.Fatalf("ThreadDelete: %v", err)
	}

	if len(q.AdminDeleteForumThreadCalls) != 1 {
		t.Fatalf("expected 1 AdminDeleteForumThread call, got %d", len(q.AdminDeleteForumThreadCalls))
	}
	if got := q.AdminDeleteForumThreadCalls[0]; got != 1 {
		t.Fatalf("expected thread id 1, got %d", got)
	}
	if len(q.SystemRebuildForumTopicMetaByIDCalls) != 1 {
		t.Fatalf("expected 1 SystemRebuildForumTopicMetaByID call, got %d", len(q.SystemRebuildForumTopicMetaByIDCalls))
	}
	if got := q.SystemRebuildForumTopicMetaByIDCalls[0]; got != 2 {
		t.Fatalf("expected topic id 2, got %d", got)
	}
}
