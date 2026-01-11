package share

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/sharesign"
)

func ShareLink(w http.ResponseWriter, r *http.Request) {
	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok || cd == nil {
		log.Printf("share link generation failed: core data missing for %s", r.URL.Path)
		http.Error(w, "core data missing", http.StatusInternalServerError)
		return
	}
	link := r.URL.Query().Get("link")
	if link == "" {
		log.Printf("share link generation failed: link parameter missing for %s", r.URL.Path)
		http.Error(w, "link is required", http.StatusBadRequest)
		return
	}
	signer, err := shareSignerFromCoreData(cd)
	if err != nil {
		log.Printf("share link generation failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if signer == nil {
		log.Printf("share link generation failed: share signer not configured for %s", r.URL.Path)
		http.Error(w, "share signer not configured", http.StatusInternalServerError)
		return
	}
	var signedURL string
	if r.URL.Query().Get("use_query") == "true" {
		signedURL = signer.SignedURLQuery(link)
	} else {
		signedURL = signer.SignedURL(link)
	}
	if signedURL == "" {
		log.Printf("share link generation failed: signed URL empty for %s (link=%s)", r.URL.Path, link)
		http.Error(w, "signed URL unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"signed_url": signedURL,
	}); err != nil {
		log.Printf("share link generation failed: response encode error for %s: %v", r.URL.Path, err)
	}
}

func shareSignerFromCoreData(cd *common.CoreData) (*sharesign.Signer, error) {
	if cd == nil {
		return nil, errors.New("core data missing")
	}
	if cd.ShareSigner != nil {
		return cd.ShareSigner, nil
	}
	if cd.Config == nil {
		return nil, errors.New("share signer missing and config unavailable")
	}
	if cd.Config.ShareSignSecret == "" {
		log.Printf("share signer missing; falling back to config share sign secret (empty secret)")
	} else {
		log.Printf("share signer missing; falling back to config share sign secret")
	}
	return sharesign.NewSigner(cd.Config, cd.Config.ShareSignSecret), nil
}
