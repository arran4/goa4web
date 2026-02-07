package faq

import (
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCustomFAQIndexPermissions(t *testing.T) {
	t.Run("Happy Path - Allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/faq", nil)

		q := testhelpers.NewQuerierStub(
			testhelpers.WithGrant("faq", "question", "post"),
		)
		ctx := req.Context()
		cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())

		CustomFAQIndex(cd, req.WithContext(ctx))
		if !common.ContainsItem(cd.CustomIndexItems, "Ask") {
			t.Errorf("expected ask item")
		}
		if len(q.SystemCheckGrantCalls) != 1 {
			t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
		}
	})

	t.Run("Unhappy Path - Denied", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/faq", nil)

		q := testhelpers.NewQuerierStub(
			testhelpers.WithDefaultGrantAllowed(false),
		)
		ctx := req.Context()
		cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())

		CustomFAQIndex(cd, req.WithContext(ctx))
		if common.ContainsItem(cd.CustomIndexItems, "Ask") {
			t.Errorf("unexpected ask item")
		}
		if len(q.SystemCheckGrantCalls) != 1 {
			t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
		}
	})
}
