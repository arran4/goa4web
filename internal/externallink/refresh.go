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
	metadataWriter
}

type metadataWriter interface {
	UpdateExternalLinkMetadata(context.Context, db.UpdateExternalLinkMetadataParams) error
	imageReferenceWriter
}

type imageReferenceWriter interface {
	UpdateExternalLinkImageCache(context.Context, db.UpdateExternalLinkImageCacheParams) error
}

// RefreshMissingImageReference repairs an incomplete cache write without refetching
// Open Graph metadata. An empty cache result leaves the optional reference unset.
func RefreshMissingImageReference(ctx context.Context, store imageReferenceWriter, link *db.ExternalLink,
	cacheImage func(context.Context, string) (string, error),
) error {
	if !link.CardImage.Valid || link.CardImage.String == "" ||
		(link.CardImageCache.Valid && link.CardImageCache.String != "") {
		return nil
	}
	ref, err := cacheImage(ctx, link.CardImage.String)
	if err != nil || ref == "" {
		return err
	}
	return storeImageReference(ctx, store, link.ID, ref)
}

func storeImageReference(ctx context.Context, store imageReferenceWriter, id int32, ref string) error {
	if err := store.UpdateExternalLinkImageCache(ctx, db.UpdateExternalLinkImageCacheParams{
		ID:             id,
		CardImageCache: sql.NullString{String: ref, Valid: true},
	}); err != nil {
		return fmt.Errorf("store card image reference: %w", err)
	}
	return nil
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
	refreshed, cacheErr, err := refreshMetadata(ctx, store, url, func(context.Context) (int32, error) {
		return id, nil
	}, fetch, cacheImage, true)
	return refreshed, errors.Join(err, cacheErr)
}

// RefreshMetadata fetches fresh metadata using the caller's client and retry
// policy. It resolves the record after fetching, so a failed fetch need not
// create a record. Image cache errors are reported separately from fetch and
// metadata storage errors.
func RefreshMetadata(ctx context.Context, store metadataWriter, url string,
	resolveID func(context.Context) (int32, error),
	fetch func(context.Context, string) (*opengraph.Info, error),
	cacheImage func(context.Context, string) (string, error),
) (cacheErr, err error) {
	_, cacheErr, err = refreshMetadata(ctx, store, url, resolveID, fetch, cacheImage, false)
	return cacheErr, err
}

func refreshMetadata(ctx context.Context, store metadataWriter, url string,
	resolveID func(context.Context) (int32, error),
	fetch func(context.Context, string) (*opengraph.Info, error),
	cacheImage func(context.Context, string) (string, error),
	writeImageOnMetadataError bool,
) (bool, error, error) {
	info, err := fetch(ctx, url)
	if err != nil {
		return false, nil, fmt.Errorf("fetch metadata: %w", err)
	}
	if info == nil {
		return false, nil, errors.New("fetch returned no metadata")
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
	id, err := resolveID(ctx)
	if err != nil || id == 0 {
		return false, cacheErr, err
	}
	normalized := *info
	normalized.Description = limitCardDescription(info.Description)
	storeErr := store.UpdateExternalLinkMetadata(ctx, metadataParams(id, &normalized))
	if storeErr != nil {
		storeErr = fmt.Errorf("store metadata: %w", storeErr)
	}
	if cachedImage != "" && (storeErr == nil || writeImageOnMetadataError) {
		cacheErr = storeImageReference(ctx, store, id, cachedImage)
	}
	return storeErr == nil, cacheErr, storeErr
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
