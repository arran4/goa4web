package faq

import (
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db/testutil"
)

func TestCustomFAQIndexAsk(t *testing.T) {
	req := httptest.NewRequest("GET", "/faq", nil)
	queries := testutil.NewBaseQuerier(t)
	queries.AllowGrants()
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	// Grant allowed via BaseQuerier.

	CustomFAQIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Ask") {
		t.Errorf("expected ask item")
	}
}

func TestCustomFAQIndexAskDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/faq", nil)
	queries := testutil.NewBaseQuerier(t)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	// Grants default to denied via BaseQuerier.

	CustomFAQIndex(cd, req.WithContext(ctx))
	if common.ContainsItem(cd.CustomIndexItems, "Ask") {
		t.Errorf("unexpected ask item")
	}
}
