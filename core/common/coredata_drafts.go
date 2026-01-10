package common

import (
	"context"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

func (cd *CoreData) CreateDraft(ctx context.Context, params db.CreateDraftParams) (int64, error) {
	if cd.Queries() == nil {
		return 0, fmt.Errorf("CreateDraft: %w", ErrDBNotInitialized)
	}
	return cd.Queries().CreateDraft(ctx, params)
}

func (cd *CoreData) UpdateDraft(ctx context.Context, params db.UpdateDraftParams) error {
	if cd.Queries() == nil {
		return fmt.Errorf("UpdateDraft: %w", ErrDBNotInitialized)
	}
	return cd.Queries().UpdateDraft(ctx, params)
}

func (cd *CoreData) GetDraft(ctx context.Context, id int32, userID int32) (*db.Draft, error) {
	if cd.Queries() == nil {
		return nil, fmt.Errorf("GetDraft: %w", ErrDBNotInitialized)
	}
	return cd.Queries().GetDraft(ctx, db.GetDraftParams{ID: id, UserID: userID})
}

func (cd *CoreData) ListDraftsForThread(ctx context.Context, threadID int32, userID int32) ([]*db.Draft, error) {
	if cd.Queries() == nil {
		return nil, fmt.Errorf("ListDraftsForThread: %w", ErrDBNotInitialized)
	}
	return cd.Queries().ListDraftsForThread(ctx, db.ListDraftsForThreadParams{ThreadID: threadID, UserID: userID})
}

func (cd *CoreData) HasDrafts(ctx context.Context, threadID int32, userID int32) (bool, error) {
	if cd.Queries() == nil {
		return false, fmt.Errorf("HasDrafts: %w", ErrDBNotInitialized)
	}
	drafts, err := cd.ListDraftsForThread(ctx, threadID, userID)
	if err != nil {
		return false, fmt.Errorf("HasDrafts: %w", err)
	}
	return len(drafts) > 0, nil
}
