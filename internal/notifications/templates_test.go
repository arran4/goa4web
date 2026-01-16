package notifications

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestRenderNotificationUsesSequentialOverrides(t *testing.T) {
	ctx := context.Background()
	q := testhelpers.NewQuerierStub()
	q.SystemGetTemplateOverrideSeq = []string{"first", "second"}
	n := New(WithQueries(q), WithConfig(config.NewRuntimeConfig()))
	filename := NotificationTemplateFilenameGenerator("reply")

	body1, err := n.renderNotification(ctx, filename, map[string]any{})
	if err != nil {
		t.Fatalf("render first: %v", err)
	}
	if string(body1) != "first" {
		t.Fatalf("expected first body, got %q", string(body1))
	}

	body2, err := n.renderNotification(ctx, filename, map[string]any{})
	if err != nil {
		t.Fatalf("render second: %v", err)
	}
	if string(body2) != "second" {
		t.Fatalf("expected second body, got %q", string(body2))
	}

	if len(q.SystemGetTemplateOverrideCalls) != 2 {
		t.Fatalf("expected 2 override lookups, got %d", len(q.SystemGetTemplateOverrideCalls))
	}
}
