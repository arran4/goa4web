package externallink

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// RedirectHandler shows a confirmation page before leaving the site or
// performs the redirect when the go parameter is present.
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	signer := cd.LinkSigner
	raw := r.URL.Query().Get("u")
	ts := r.URL.Query().Get("ts")
	sig := r.URL.Query().Get("sig")
	if signer == nil || raw == "" || !signer.Verify(raw, ts, sig) {
		http.Error(w, "invalid link", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("go") != "" {
		if err := cd.Queries().RegisterExternalLinkClick(r.Context(), raw); err != nil {
			log.Printf("record external link click: %v", err)
		}
		http.Redirect(w, r, raw, http.StatusTemporaryRedirect)
		return
	}

	link, err := cd.Queries().GetExternalLink(r.Context(), raw)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("get external link: %v", err)
	}

	var cardImgURL, faviconURL string
	if link != nil {
		if link.CardImageCache.Valid && cd.ImageSigner != nil {
			cardImgURL = cd.ImageSigner.SignedCacheURL(link.CardImageCache.String)
		}
		if link.FaviconCache.Valid && cd.ImageSigner != nil {
			faviconURL = cd.ImageSigner.SignedCacheURL(link.FaviconCache.String)
		}
	}

	type Data struct {
		*common.CoreData
		URL         string
		RedirectURL string
		Title       string
		Desc        string
		CardImage   string
		Favicon     string
		ReloadURL   string
	}
	cd.PageTitle = "External Link"
	reloadURL := fmt.Sprintf("/goto?u=%s&ts=%s&sig=%s&reload=1", url.QueryEscape(raw), ts, sig)
	data := Data{
		CoreData:    cd,
		URL:         raw,
		RedirectURL: fmt.Sprintf("/goto?u=%s&ts=%s&sig=%s&go=1", url.QueryEscape(raw), ts, sig),
		ReloadURL:   reloadURL,
	}
	if link != nil {
		if link.CardTitle.Valid {
			data.Title = link.CardTitle.String
		}
		if link.CardDescription.Valid {
			data.Desc = link.CardDescription.String
		}
		if cardImgURL != "" {
			data.CardImage = cardImgURL
		} else if link.CardImage.Valid {
			data.CardImage = link.CardImage.String
		}
		if faviconURL != "" {
			data.Favicon = faviconURL
		}
	}
	if err := cd.ExecuteSiteTemplate(w, r, "externalLinkPage.gohtml", data); err != nil {
		log.Printf("Template Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
