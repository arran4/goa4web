package writings

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestWritingsFeed(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}
	now := time.Now()
	q.GetPublicWritingsReturns = []*db.Writing{
		{
			Idwriting: 1,
			Title:     sql.NullString{String: "Title", Valid: true},
			Published: sql.NullTime{Time: now, Valid: true},
			Writing:   sql.NullString{String: "Content", Valid: true},
			Abstract:  sql.NullString{String: "Abstract", Valid: true},
		},
	}

	req := httptest.NewRequest("GET", "http://example.com/writings/rss", nil)
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSiteTitle("Site"))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	feed, err := feedGen(req, cd)
	if err != nil {
		t.Fatalf("feedGen: %v", err)
	}

	if feed.Title != "Site - Latest writings" {
		t.Errorf("feed title incorrect: %s", feed.Title)
	}
	if len(feed.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(feed.Items))
	}
}
