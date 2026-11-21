package websocket

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/arran4/goa4web/core/common"
)

// LiveEventKindExternalLinkCard asks the browser to refresh its visible card fragments.
const LiveEventKindExternalLinkCard = "external-link-card"

// linkCardPollInterval bounds database reads while waiting for an independent worker.
const linkCardPollInterval = 3 * time.Second

// maxLiveCards bounds the number of source links observed per connection.
const maxLiveCards = 12

type linkCardWatch struct {
	viewer   *common.CoreData
	threadID int32
	seen     map[string]common.ForumLinkCard
}

func watchLinkCards(ctx context.Context, viewer *common.CoreData, threadID int32, keys []string) *linkCardWatch {
	watch := &linkCardWatch{
		viewer: viewer, threadID: threadID,
		seen: make(map[string]common.ForumLinkCard),
	}
	for _, key := range keys[:min(len(keys), maxLiveCards)] {
		if _, duplicate := watch.seen[key]; duplicate {
			continue
		}
		if card, err := viewer.ForumLinkCardForViewer(ctx, threadID, key); err == nil {
			watch.seen[key] = card
		}
	}
	if len(watch.seen) == 0 {
		return nil
	}
	return watch
}

func (watch *linkCardWatch) changed(ctx context.Context) bool {
	var changed bool
	for key, previous := range watch.seen {
		current, err := watch.viewer.StoredLinkCard(ctx, previous.URL)
		if err != nil || current == previous {
			continue
		}
		// Re-resolve the URL from the visible source before announcing any change.
		card, err := watch.viewer.ForumLinkCardForViewer(ctx, watch.threadID, key)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			delete(watch.seen, key)
		case err == nil && card != previous:
			watch.seen[key] = card
			changed = true
		}
	}
	return changed
}
