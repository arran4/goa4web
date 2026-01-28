package handlertest

import (
	"context"
	"net/http"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

// NewCoreData creates a CoreData configured with a fake Querier stub.
func NewCoreData(t *testing.T, ctx context.Context, opts ...common.CoreOption) (*common.CoreData, *db.QuerierStub, func()) {
	t.Helper()
	stub := testhelpers.NewQuerierStub()
	cd := common.NewCoreData(ctx, stub, config.NewRuntimeConfig(), opts...)
	return cd, stub, func() {}
}

// RequestWithCoreData attaches CoreData with a fake Querier stub to req's context.
// The returned request includes the CoreData under consts.KeyCoreData.
func RequestWithCoreData(t *testing.T, req *http.Request, opts ...common.CoreOption) (*http.Request, *common.CoreData, *db.QuerierStub, func()) {
	t.Helper()
	cd, stub, cleanup := NewCoreData(t, req.Context(), opts...)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	return req.WithContext(ctx), cd, stub, cleanup
}
