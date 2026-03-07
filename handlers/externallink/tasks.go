package externallink

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/opengraph"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// ReloadExternalLinkTask reloads OG metadata for a link.
type ReloadExternalLinkTask struct{ tasks.TaskString }

var reloadExternalLinkTask = &ReloadExternalLinkTask{TaskString: "admin:externallink:reload"}

// ensure conformance
var _ tasks.Task = (*ReloadExternalLinkTask)(nil)

func (ReloadExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd.LinkSignKey == "" {
		return fmt.Errorf("invalid link config")
	}

	u := r.FormValue("u")
	id := r.FormValue("id")
	sig := r.FormValue("sig")

	data := ""
	if u != "" {
		data = "link:" + u
	} else if id != "" {
		data = "link:" + id
	} else {
		return fmt.Errorf("missing u or id")
	}

	if err := sign.Verify(data, sig, cd.LinkSignKey, sign.WithOutNonce()); err != nil {
		return fmt.Errorf("invalid signature: %w", handlers.ErrForbidden)
	}

	rawURL := u
	if id != "" {
		// If ID is provided, we should look it up to get the URL
		lid := int32(0)
		fmt.Sscan(id, &lid)
		if lid != 0 {
			if l, err := cd.Queries().GetExternalLinkByID(r.Context(), lid); err == nil {
				rawURL = l.Url
			}
		}
	}

	if rawURL == "" {
		return fmt.Errorf("no url provided")
	}

	// Spawn a goroutine to fetch OpenGraph data in the background
	go func() {
		// create a disconnected context or use background context for DB operations
		// since the request context will be canceled when this handler returns.
		bgCtx := context.Background()
		info, err := opengraph.Fetch(rawURL, cd.HTTPClient())
		if err != nil {
			log.Printf("background fetch error for %s: %v", rawURL, err)
			return
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
		res, err := cd.EnsureExternalLink(bgCtx, rawURL)
		var lid int32
		if err == nil {
			id, _ := res.LastInsertId()
			lid = int32(id)
		} else {
			// Fallback to fetch existing if EnsureExternalLink fails for any reason
			l, err := cd.GetExternalLink(bgCtx, rawURL)
			if err == nil {
				lid = l.ID
			}
		}

		if lid != 0 {
			err := cd.UpdateExternalLinkMetadata(bgCtx, db.UpdateExternalLinkMetadataParams{
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
				return
			}

			if cachedImgName != "" {
				// Update cache
				err := cd.UpdateExternalLinkImageCache(bgCtx, db.UpdateExternalLinkImageCacheParams{
					CardImageCache: sql.NullString{String: cachedImgName, Valid: true},
					ID:             lid,
				})
				if err != nil {
					// non-fatal, just log
					log.Printf("failed to update cache: %v", err)
				}
			}
		}
	}()

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

func (t *ReloadExternalLinkTask) Matcher() func(*http.Request, *mux.RouteMatch) bool {
	return func(r *http.Request, rm *mux.RouteMatch) bool {
		return true
	}
}
