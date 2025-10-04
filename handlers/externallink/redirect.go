package externallink

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// RedirectHandler shows a confirmation page before leaving the site or
// performs the redirect when the go parameter is present.
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	signer := cd.LinkSigner
	idStr := r.URL.Query().Get("id")
	ts := r.URL.Query().Get("ts")
	sig := r.URL.Query().Get("sig")
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if signer == nil || idStr == "" || err != nil || !signer.Verify(idStr, ts, sig) {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
		return
	}
	id := int32(id64)
	cd.SetCurrentExternalLinkID(id)
	link := cd.SelectedExternalLink()
	if link == nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
		return
	}
	raw := link.Url
	if r.URL.Query().Get("go") != "" {
		cd.RegisterExternalLinkClick(raw)
		handlers.RedirectToGet(w, r, raw)
		return
	}

	type Data struct {
		URL         string
		RedirectURL string
		ReloadURL   string
	}
	cd.PageTitle = "External Link"
	data := Data{
		URL:         raw,
		RedirectURL: fmt.Sprintf("/goto?id=%s&ts=%s&sig=%s&go=1", idStr, ts, sig),
		ReloadURL:   fmt.Sprintf("/goto?id=%s&ts=%s&sig=%s&reload=1", idStr, ts, sig),
	}
	if err := cd.ExecuteSiteTemplate(w, r, "externalLinkPage.gohtml", data); err != nil {
		log.Printf("Template Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
	}
}
