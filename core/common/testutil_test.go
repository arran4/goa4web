package common

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

// QuerierStub wraps a db.Querier for tests that need CoreData with stubbed queries.
type QuerierStub struct {
	db.Querier
}

// NewTestCoreData returns a CoreData configured for tests using a QuerierStub.
// Use WithUserRoles or other CoreOption helpers to override defaults.
func NewTestCoreData(t *testing.T, q db.Querier) *CoreData {
	t.Helper()
	if q == nil {
		return NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	}
	return NewCoreData(context.Background(), q, config.NewRuntimeConfig())
}
