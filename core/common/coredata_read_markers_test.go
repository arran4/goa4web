package common_test

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/db/testutil"
)

func TestThreadReadMarker(t *testing.T) {
	queries := testutil.NewReadMarkersQuerier(t)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 1

	if err := cd.SetThreadReadMarker(2, 5); err != nil {
		t.Fatalf("SetThreadReadMarker: %v", err)
	}

	if len(queries.Upserted) != 1 {
		t.Fatalf("expected one upsert, got %d", len(queries.Upserted))
	}
	arg := queries.Upserted[0]
	if arg.Item != "thread" || arg.ItemID != 2 || arg.UserID != cd.UserID || arg.LastCommentID != 5 {
		t.Fatalf("unexpected read marker args: %+v", arg)
	}

	queries.Marker = &db.GetContentReadMarkerRow{Item: "thread", ItemID: 2, UserID: cd.UserID, LastCommentID: 5}

	cid, err := cd.ThreadReadMarker(2)
	if err != nil {
		t.Fatalf("ThreadReadMarker: %v", err)
	}
	if cid != 5 {
		t.Fatalf("last comment %d, want 5", cid)
	}
}
