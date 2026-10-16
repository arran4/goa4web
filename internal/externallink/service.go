// Package externallink owns external-link persistence decisions shared by adapters.
package externallink

import (
	"context"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// ResolveRecord ensures url has an external-link record and returns its ID.
func ResolveRecord(ctx context.Context, q db.Querier, url string) (int32, error) {
	result, err := q.EnsureExternalLink(ctx, url)
	if err == nil {
		id, idErr := result.LastInsertId()
		if idErr == nil && id > 0 {
			return int32(id), nil
		}
	}

	link, lookupErr := q.GetExternalLink(ctx, url)
	if lookupErr != nil {
		if err != nil {
			return 0, fmt.Errorf("ensure external link: %w", err)
		}
		return 0, fmt.Errorf("lookup external link: %w", lookupErr)
	}
	if link == nil || link.ID == 0 {
		return 0, nil
	}
	return link.ID, nil
}

// NeedsMetadata reports whether link has not yet received Open Graph metadata.
func NeedsMetadata(link *db.ExternalLink) bool {
	return link == nil || !link.CardTitle.Valid || link.CardTitle.String == ""
}
