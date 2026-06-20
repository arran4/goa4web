package externallink

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/opengraph"
)

// RedirectHandler shows a confirmation page before leaving the site or
// performs the redirect when the go parameter is present.
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	rawURL, linkID, existingLink, usedURL, err := cd.ResolveExternalLink(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
		return
	}

	if r.URL.Query().Get("go") != "" {
		cd.RegisterExternalLinkClick(rawURL)
		http.Redirect(w, r, rawURL, http.StatusTemporaryRedirect)
		return
	}

	// Check if the URL is internal to avoid self-fetching
	isInternal := false
	if u, err := url.Parse(rawURL); err == nil {
		if u.Hostname() == r.Host {
			isInternal = true
		}
		// Also check against configured hostname
		if cd.Config != nil && cd.Config.BaseURL != "" {
			if confU, err := url.Parse(cd.Config.BaseURL); err == nil {
				if u.Hostname() == confU.Hostname() {
					isInternal = true
				}
			}
		}
	}

	needFetch := false
	if linkID == 0 && usedURL {
		needFetch = true
	} else if existingLink != nil && (!existingLink.CardTitle.Valid || !existingLink.CardDescription.Valid) {
		needFetch = true
	}

	if needFetch && rawURL != "" && !isInternal {
		info, err := opengraph.Fetch(rawURL, cd.HTTPClient())
		if err == nil {
			if linkID == 0 {
			res, err := cd.EnsureExternalLink(r.Context(), rawURL)
				if err == nil {
					id, _ := res.LastInsertId()
					linkID = int32(id)
				} else {
				log.Printf("EnsureExternalLink error: %v", err)
				}
			}
			if linkID != 0 {
				err := cd.UpdateExternalLinkMetadata(r.Context(), db.UpdateExternalLinkMetadataParams{
					CardTitle:       sql.NullString{String: info.Title, Valid: info.Title != ""},
					CardDescription: sql.NullString{String: info.Description, Valid: info.Description != ""},
					CardImage:       sql.NullString{String: info.Image, Valid: info.Image != ""},
					CardDuration:    sql.NullString{String: info.Duration, Valid: info.Duration != ""},
					CardUploadDate:  sql.NullString{String: info.UploadDate, Valid: info.UploadDate != ""},
					CardAuthor:      sql.NullString{String: info.Author, Valid: info.Author != ""},
					ID:              linkID,
				})
				if err != nil {
					log.Printf("UpdateExternalLinkMetadata error: %v", err)
				}
			}
		} else {
			log.Printf("fetchOpenGraph error: %v", err)
		}
	} else if isInternal {
		// Log or handle internal link specifically if needed
		// For now, we just don't fetch metadata
	}

	link := cd.SelectedExternalLink()
	if link != nil && link.CardImage.Valid && !link.CardImageCache.Valid {
		cached, err := cd.DownloadAndCacheImage(link.CardImage.String)
		if err == nil {
			_ = cd.UpdateExternalLinkImageCache(r.Context(), db.UpdateExternalLinkImageCacheParams{
				CardImageCache: sql.NullString{String: cached, Valid: true},
				ID:             link.ID,
			})
			link.CardImageCache = sql.NullString{String: cached, Valid: true}
		}
	}

	type Data struct {
		Message     string
		BackURL     string
	}
	cd.PageTitle = "External Link"
	data := Data{
		Message:     r.URL.Query().Get("msg"),
		BackURL:     r.Referer(),
	}
	if err := cd.ExecuteSiteTemplate(w, r, "externalLinkPage.gohtml", data); err != nil {
		log.Printf("Template Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
	}
}
