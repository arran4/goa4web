package externallink

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/opengraph"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// ReloadExternalLinkTask reloads OG metadata for a link.
type ReloadExternalLinkTask struct {
	tasks.TaskString
	Config *config.RuntimeConfig
	RawURL string
}

var reloadExternalLinkTask = ReloadExternalLinkTask{TaskString: "admin:externallink:reload"}

// ensure conformance
var _ tasks.Task = ReloadExternalLinkTask{}
var _ tasks.BackgroundTasker = ReloadExternalLinkTask{}

func (t ReloadExternalLinkTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	cd := common.NewCoreData(ctx, q, t.Config)
	info, err := opengraph.Fetch(t.RawURL, cd.HTTPClient())
	if err != nil {
		log.Printf("background fetch error for %s: %v", t.RawURL, err)
		return nil, err
	}

	var cachedImgName string
	if info.Image != "" {
		var err error
		cachedImgName, err = cd.DownloadAndCacheImage(info.Image)
		if err != nil {
			log.Printf("failed to cache image: %v", err)
		}
	}

	// Update DB using EnsureExternalLink to handle duplicates properly
	res, err := cd.EnsureExternalLink(ctx, t.RawURL)
	var lid int32
	if err == nil {
		id, _ := res.LastInsertId()
		lid = int32(id)
	}

	// Always fetch existing if EnsureExternalLink returns 0 or fails
	if lid == 0 {
		if l, err := cd.GetExternalLink(ctx, t.RawURL); err == nil && l != nil {
			lid = l.ID
		}
	}

	if lid != 0 {
		err := cd.UpdateExternalLinkMetadata(ctx, db.UpdateExternalLinkMetadataParams{
			CardTitle:       sql.NullString{String: info.Title, Valid: info.Title != ""},
			CardDescription: sql.NullString{String: info.Description, Valid: info.Description != ""},
			CardImage:       sql.NullString{String: info.Image, Valid: info.Image != ""},
			CardDuration:    sql.NullString{String: info.Duration, Valid: info.Duration != ""},
			CardUploadDate:  sql.NullString{String: info.UploadDate, Valid: info.UploadDate != ""},
			CardAuthor:      sql.NullString{String: info.Author, Valid: info.Author != ""},
			ID:              lid,
		})
		if err != nil {
			log.Printf("background update error: %v", err)
			return nil, err
		}

		if cachedImgName != "" {
			// Update cache
			err := cd.UpdateExternalLinkImageCache(ctx, db.UpdateExternalLinkImageCacheParams{
				CardImageCache: sql.NullString{String: cachedImgName, Valid: true},
				ID:             lid,
			})
			if err != nil {
				// non-fatal, just log
				log.Printf("failed to update cache: %v", err)
			}
		}
	}
	return nil, nil
}

func (t ReloadExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	rawURL, _, _, _, err := cd.ResolveExternalLink(r)
	if err != nil {
		return fmt.Errorf("invalid link: %w", handlers.ErrForbidden)
	}

	cd.SetEventTask(ReloadExternalLinkTask{
		TaskString: t.TaskString,
		Config:     cd.Config,
		RawURL:     rawURL,
	})

	// Return redirect to the current URL with msg parameter
	// We reconstruct the URL from params to be safe, or just use RequestURI?
	// RequestURI includes the path and query.
	// Since we are POSTing to /goto?u=... , RequestURI is exactly what we want to GET.
	redirectURI := r.RequestURI
	if r.URL.Query().Get("msg") == "" {
		redirectURI += "&msg=Reloading+Open+Graph+data+in+the+background..."
	}
	return handlers.RedirectHandler(redirectURI)
}

func (t ReloadExternalLinkTask) Matcher() func(*http.Request, *mux.RouteMatch) bool {
	return func(r *http.Request, rm *mux.RouteMatch) bool {
		return true
	}
}

// PrefetchExternalLinkTask prefetches OG metadata for a link.
type PrefetchExternalLinkTask struct {
	tasks.TaskString
	Config *config.RuntimeConfig
	RawURL string
}

var prefetchExternalLinkTask = PrefetchExternalLinkTask{TaskString: "externallink:prefetch"}

// ensure conformance
var _ tasks.Task = PrefetchExternalLinkTask{}
var _ tasks.BackgroundTasker = PrefetchExternalLinkTask{}

func (t PrefetchExternalLinkTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	cd := common.NewCoreData(ctx, q, t.Config)
	info, err := opengraph.Fetch(t.RawURL, cd.HTTPClient())
	if err != nil {
		log.Printf("background fetch error for %s: %v", t.RawURL, err)
		return nil, err
	}

	var cachedImgName string
	if info.Image != "" {
		var err error
		cachedImgName, err = cd.DownloadAndCacheImage(info.Image)
		if err != nil {
			log.Printf("failed to cache image: %v", err)
		}
	}

	// Update DB using EnsureExternalLink to handle duplicates properly
	res, err := cd.EnsureExternalLink(ctx, t.RawURL)
	var lid int32
	if err == nil {
		id, _ := res.LastInsertId()
		lid = int32(id)
	}

	// Always fetch existing if EnsureExternalLink returns 0 or fails
	if lid == 0 {
		if l, err := cd.GetExternalLink(ctx, t.RawURL); err == nil && l != nil {
			lid = l.ID
		}
	}

	if lid != 0 {
		err := cd.UpdateExternalLinkMetadata(ctx, db.UpdateExternalLinkMetadataParams{
			CardTitle:       sql.NullString{String: info.Title, Valid: info.Title != ""},
			CardDescription: sql.NullString{String: info.Description, Valid: info.Description != ""},
			CardImage:       sql.NullString{String: info.Image, Valid: info.Image != ""},
			CardDuration:    sql.NullString{String: info.Duration, Valid: info.Duration != ""},
			CardUploadDate:  sql.NullString{String: info.UploadDate, Valid: info.UploadDate != ""},
			CardAuthor:      sql.NullString{String: info.Author, Valid: info.Author != ""},
			ID:              lid,
		})
		if err != nil {
			log.Printf("background update error: %v", err)
			return nil, err
		}

		if cachedImgName != "" {
			// Update cache
			err := cd.UpdateExternalLinkImageCache(ctx, db.UpdateExternalLinkImageCacheParams{
				CardImageCache: sql.NullString{String: cachedImgName, Valid: true},
				ID:             lid,
			})
			if err != nil {
				// non-fatal, just log
				log.Printf("failed to update cache: %v", err)
			}
		}
	}
	return nil, nil
}

func (t PrefetchExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	rawURL := r.FormValue("url")
	if rawURL == "" {
		return handlers.TextByteWriter("missing url")
	}

	cd.SetEventTask(PrefetchExternalLinkTask{
		TaskString: t.TaskString,
		Config:     cd.Config,
		RawURL:     rawURL,
	})

	return handlers.TextByteWriter("ok")
}
