package externallink

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/opengraph"
)

type metadataStore interface {
	GetExternalLinkByID(context.Context, int32) (*db.ExternalLink, error)
	UpdateExternalLinkMetadata(context.Context, db.UpdateExternalLinkMetadataParams) error
	UpdateExternalLinkImageCache(context.Context, db.UpdateExternalLinkImageCacheParams) error
}

// RefreshMissingMetadata fetches and persists metadata for a link without a title.
// The fetch adapter supplies its retry and client policy. A cache failure is
// returned separately from the refreshed result so the metadata can still be used.
func RefreshMissingMetadata(ctx context.Context, store metadataStore, id int32, url string,
	fetch func(context.Context, string) (*opengraph.Info, error),
	cacheImage func(context.Context, string) (string, error),
) (refreshed bool, err error) {
	link, err := store.GetExternalLinkByID(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("load metadata state: %w", err)
	}
	if !NeedsMetadata(link) {
		return false, nil
	}
	info, err := fetch(ctx, url)
	if err != nil {
		return false, fmt.Errorf("fetch metadata: %w", err)
	}
	if info == nil {
		return false, errors.New("fetch returned no metadata")
	}

	var cachedImage string
	var cacheErr error
	if cacheImage != nil && info.Image != "" {
		cachedImage, err = cacheImage(ctx, info.Image)
		if err != nil {
			cachedImage = ""
			cacheErr = fmt.Errorf("cache card image: %w", err)
		}
	}
	normalized := *info
	normalized.Description = limitCardDescription(info.Description)
	storeErr := store.UpdateExternalLinkMetadata(ctx, MetadataParams(id, &normalized))
	if storeErr != nil {
		storeErr = fmt.Errorf("store metadata: %w", storeErr)
	}
	if cachedImage != "" {
		err := store.UpdateExternalLinkImageCache(ctx, db.UpdateExternalLinkImageCacheParams{
			ID:             id,
			CardImageCache: sql.NullString{String: cachedImage, Valid: true},
		})
		if err != nil {
			cacheErr = fmt.Errorf("store card image reference: %w", err)
		}
	}
	return storeErr == nil, errors.Join(storeErr, cacheErr)
}

// limitCardDescription retains the worker's existing byte threshold and rune cap.
func limitCardDescription(desc string) string {
	const maxDescriptionBytes = 65000       // Leave space within the SQL TEXT limit.
	const truncatedDescriptionRunes = 16000 // At most 64000 UTF-8 bytes.
	if len(desc) <= maxDescriptionBytes {
		return desc
	}
	var count int
	for offset := range desc {
		if count == truncatedDescriptionRunes {
			return desc[:offset]
		}
		count++
	}
	return desc
}
