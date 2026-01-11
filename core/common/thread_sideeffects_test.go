package common

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

func TestHandleThreadUpdatedMarksThreadAndItemLabels(t *testing.T) {
	ctx := context.Background()
	queries := &db.QuerierStub{}
	cd := NewCoreData(ctx, queries, config.NewRuntimeConfig())
	cd.UserID = 5

	err := cd.HandleThreadUpdated(ctx, ThreadUpdatedEvent{
		ThreadID:             12,
		CommentID:            7,
		LabelItem:            "news",
		LabelItemID:          99,
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
	})
	if err != nil {
		t.Fatalf("HandleThreadUpdated: %v", err)
	}

	if len(queries.ClearUnreadContentPrivateLabelExceptUserCalls) != 2 {
		t.Fatalf("expected 2 clear unread calls, got %d", len(queries.ClearUnreadContentPrivateLabelExceptUserCalls))
	}
	seenClear := map[string]int32{}
	for _, call := range queries.ClearUnreadContentPrivateLabelExceptUserCalls {
		seenClear[call.Item] = call.ItemID
	}
	if seenClear["thread"] != 12 {
		t.Fatalf("expected thread unread clear for 12, got %d", seenClear["thread"])
	}
	if seenClear["news"] != 99 {
		t.Fatalf("expected news unread clear for 99, got %d", seenClear["news"])
	}

	if len(queries.UpsertContentReadMarkerCalls) != 1 {
		t.Fatalf("expected 1 read marker call, got %d", len(queries.UpsertContentReadMarkerCalls))
	}
	if got := queries.UpsertContentReadMarkerCalls[0]; got.Item != "thread" || got.ItemID != 12 {
		t.Fatalf("unexpected read marker call: %+v", got)
	}

	if len(queries.AddContentPrivateLabelCalls) != 4 {
		t.Fatalf("expected 4 label upserts, got %d", len(queries.AddContentPrivateLabelCalls))
	}
	seenLabels := map[string]map[string]bool{}
	for _, call := range queries.AddContentPrivateLabelCalls {
		if seenLabels[call.Item] == nil {
			seenLabels[call.Item] = map[string]bool{}
		}
		seenLabels[call.Item][call.Label] = call.Invert
	}
	for _, item := range []string{"thread", "news"} {
		if labels := seenLabels[item]; !labels["new"] || !labels["unread"] {
			t.Fatalf("expected %s new/unread labels inverted, got %+v", item, labels)
		}
	}
}
