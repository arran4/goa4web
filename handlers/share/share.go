package share

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/sharesign"
)

func ShareLink(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	link := r.URL.Query().Get("link")
	if link == "" {
		http.Error(w, "link is required", http.StatusBadRequest)
		return
	}
	log.Printf("Generating share link for: %s, UseQuery: %s", link, r.URL.Query().Get("use_query"))
	signer := sharesign.NewSigner(cd.Config, cd.Config.ShareSignSecret)
	var signedURL string
	if r.URL.Query().Get("use_query") == "true" {
		signedURL = signer.SignedURLQuery(link)
	} else {
		signedURL = signer.SignedURL(link)
	}
	log.Printf("Generated signed URL: %s", signedURL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"signed_url": signedURL,
	})
}
