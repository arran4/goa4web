package externallink

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/sign"
)

// RedirectHandler shows a confirmation page before leaving the site or
// performs the redirect when the go parameter is present.
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd.LinkSignKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
		return
	}
	idStr := r.URL.Query().Get("id")
	rawURL := r.URL.Query().Get("u")
	sig := r.URL.Query().Get("sig")
	var linkID int32
	usedURL := false
	switch {
	case rawURL != "":
		data := "link:" + rawURL
		if err := sign.Verify(data, sig, cd.LinkSignKey, sign.WithOutNonce()); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
			return
		}
		usedURL = true
		if queries := cd.Queries(); queries != nil {
			if link, err := queries.GetExternalLink(r.Context(), rawURL); err == nil && link != nil {
				linkID = link.ID
			} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Printf("load external link by url: %v", err)
			}
		}
	case idStr != "":
		id64, err := strconv.ParseInt(idStr, 10, 32)
		data := "link:" + idStr
		if err != nil || sign.Verify(data, sig, cd.LinkSignKey, sign.WithOutNonce()) != nil {
			w.WriteHeader(http.StatusBadRequest)
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
			return
		}
		linkID = int32(id64)
	default:
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
		return
	}
	if linkID != 0 {
		cd.SetCurrentExternalLinkID(linkID)
	}
	link := cd.SelectedExternalLink()
	if rawURL == "" {
		if link == nil {
			w.WriteHeader(http.StatusBadRequest)
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
			return
		}
		rawURL = link.Url
	}
	if r.URL.Query().Get("go") != "" {
		cd.RegisterExternalLinkClick(rawURL)
		http.Redirect(w, r, rawURL, http.StatusTemporaryRedirect)
		return
	}

	type Data struct {
		URL         string
		RedirectURL string
		ReloadURL   string
	}
	cd.PageTitle = "External Link"
	linkParam := "id"
	linkValue := idStr
	if usedURL {
		linkParam = "u"
		linkValue = url.QueryEscape(rawURL)
	}
	data := Data{
		URL:         rawURL,
		RedirectURL: fmt.Sprintf("/goto?%s=%s&sig=%s&go=1", linkParam, linkValue, sig),
		ReloadURL:   fmt.Sprintf("/goto?%s=%s&sig=%s&reload=1", linkParam, linkValue, sig),
	}
	if err := cd.ExecuteSiteTemplate(w, r, "externalLinkPage.gohtml", data); err != nil {
		log.Printf("Template Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
	}
}
