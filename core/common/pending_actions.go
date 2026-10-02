package common

import (
	"context"

	"github.com/arran4/goa4web/internal/db"
)

// CreatePendingAction stores a server-side form nonce or resumable action.
func (cd *CoreData) CreatePendingAction(ctx context.Context, params db.SystemInsertPendingActionParams) error {
	return cd.queries.SystemInsertPendingAction(ctx, params)
}

// PendingAction returns one live, unconsumed pending action by hashed token.
func (cd *CoreData) PendingAction(ctx context.Context, tokenHash string) (*db.PendingAction, error) {
	return cd.queries.SystemGetPendingAction(ctx, tokenHash)
}

// CapturePendingAction atomically replaces a form nonce with a resume token and payload.
func (cd *CoreData) CapturePendingAction(ctx context.Context, params db.SystemCapturePendingActionParams) (int64, error) {
	return cd.queries.SystemCapturePendingAction(ctx, params)
}

// ConsumePendingAction atomically retires a live pending action.
func (cd *CoreData) ConsumePendingAction(ctx context.Context, tokenHash string) (int64, error) {
	return cd.queries.SystemConsumePendingAction(ctx, tokenHash)
}

// DeleteExpiredPendingActions removes expired form and resume records.
func (cd *CoreData) DeleteExpiredPendingActions(ctx context.Context) (int64, error) {
	return cd.queries.SystemDeleteExpiredPendingActions(ctx)
}
