package share

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arran4/goa4web/internal/sharesign"
)

type ShareHandler struct {
	signer *sharesign.Signer
}

func NewShareHandler(signer *sharesign.Signer) *ShareHandler {
	return &ShareHandler{signer: signer}
}

func (h *ShareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Query().Get("link")
	if link == "" {
		http.Error(w, "link parameter is required", http.StatusBadRequest)
		return
	}
	log.Printf("Generating share link (Handler) for: %s, UseQuery: %s", link, r.URL.Query().Get("use_query"))

	var signedURL string
	var err error
	if r.URL.Query().Get("use_query") == "true" {
		signedURL, err = h.signer.SignedURLQuery(link)
	} else {
		signedURL, err = h.signer.SignedURL(link)
	}
	if err != nil {
		log.Printf("[ShareHandler] Error signing URL: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Generated signed URL (Handler): %s", signedURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"signed_url": signedURL,
	})
}
