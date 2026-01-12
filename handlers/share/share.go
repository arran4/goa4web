package share

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func ShareLink(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	link := r.URL.Query().Get("link")
	if link == "" {
		http.Error(w, "link is required", http.StatusBadRequest)
		return
	}
	log.Printf("Generating share link for: %s, UseQuery: %s", link, r.URL.Query().Get("use_query"))

	var signedURL string
	var err error
	if r.URL.Query().Get("use_query") == "true" {
		signedURL, err = cd.SignShareURLQuery(link)
	} else {
		signedURL, err = cd.SignShareURL(link)
	}
	if err != nil {
		log.Printf("[Share] Error signing URL: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Generated signed URL: %s", signedURL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"signed_url": signedURL,
	})
}
