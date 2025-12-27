package handlertest

import (
	"context"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

// NewCoreData creates a CoreData configured with a fake Querier backed by sqlmock.
// A cleanup function is returned to close the mock database when the test finishes.
func NewCoreData(t *testing.T, ctx context.Context, opts ...common.CoreOption) (*common.CoreData, sqlmock.Sqlmock, func()) {
	t.Helper()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	cd := common.NewCoreData(ctx, db.New(conn), config.NewRuntimeConfig(), opts...)
	return cd, mock, func() { conn.Close() }
}

// RequestWithCoreData attaches CoreData with a fake Querier to req's context.
// The returned request includes the CoreData under consts.KeyCoreData.
func RequestWithCoreData(t *testing.T, req *http.Request, opts ...common.CoreOption) (*http.Request, *common.CoreData, sqlmock.Sqlmock, func()) {
	t.Helper()
	cd, mock, cleanup := NewCoreData(t, req.Context(), opts...)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	return req.WithContext(ctx), cd, mock, cleanup
}
