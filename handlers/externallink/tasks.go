package externallink

import (
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

	title, desc, imgURL, err := opengraph.Fetch(rawURL, cd.HTTPClient())
	if err != nil {
		return fmt.Errorf("fetch error: %w", err)
	}

	var cachedImgName string
	if imgURL != "" {
		var err error
		cachedImgName, err = cd.DownloadAndCacheImage(imgURL)
		if err != nil {
			log.Printf("failed to cache image: %v", err)
		}
	}

	// Update DB
	res, err := cd.Queries().CreateExternalLink(r.Context(), rawURL)
	var lid int32
	if err == nil {
		id, _ := res.LastInsertId()
		lid = int32(id)
	} else {
		// Try to find existing
		l, err := cd.Queries().GetExternalLink(r.Context(), rawURL)
		if err == nil {
			lid = l.ID
		}
	}

	if lid != 0 {
		err := cd.Queries().UpdateExternalLinkMetadata(r.Context(), db.UpdateExternalLinkMetadataParams{
			CardTitle:       sql.NullString{String: title, Valid: title != ""},
			CardDescription: sql.NullString{String: desc, Valid: desc != ""},
			CardImage:       sql.NullString{String: imgURL, Valid: imgURL != ""},
			ID:              lid,
		})
		if err != nil {
			return fmt.Errorf("update error: %w", err)
		}

		if cachedImgName != "" {
			// Update cache
			err := cd.Queries().UpdateExternalLinkImageCache(r.Context(), db.UpdateExternalLinkImageCacheParams{
				CardImageCache: sql.NullString{String: cachedImgName, Valid: true},
				ID:             lid,
			})
			if err != nil {
				// non-fatal, just log
				log.Printf("failed to update cache: %v", err)
			}
		}
	}

	// Return redirect to the current URL
	// We reconstruct the URL from params to be safe, or just use RequestURI?
	// RequestURI includes the path and query.
	// Since we are POSTing to /goto?u=... , RequestURI is exactly what we want to GET.
	return handlers.RedirectHandler(r.RequestURI)
}

func (t *ReloadExternalLinkTask) Matcher() func(*http.Request, *mux.RouteMatch) bool {
	return func(r *http.Request, rm *mux.RouteMatch) bool {
		return true
	}
}
