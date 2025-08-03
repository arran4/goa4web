package externallink

import (
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
		cd.RegisterExternalLinkClick(raw)
		http.Redirect(w, r, raw, http.StatusTemporaryRedirect)
		return
	}

	type Data struct {
		URL         string
		RedirectURL string
		ReloadURL   string
	}
	cd.PageTitle = "External Link"
	reloadURL := fmt.Sprintf("/goto?u=%s&ts=%s&sig=%s&reload=1", url.QueryEscape(raw), ts, sig)
	data := Data{
		URL:         raw,
		RedirectURL: fmt.Sprintf("/goto?u=%s&ts=%s&sig=%s&go=1", url.QueryEscape(raw), ts, sig),
		ReloadURL:   reloadURL,
	}
	if err := cd.ExecuteSiteTemplate(w, r, "externalLinkPage.gohtml", data); err != nil {
		log.Printf("Template Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
