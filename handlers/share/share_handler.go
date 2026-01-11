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
		log.Printf("share link generation failed: link parameter missing for %s", r.URL.Path)
		http.Error(w, "link parameter is required", http.StatusBadRequest)
		return
	}

	var signedURL string
	if r.URL.Query().Get("use_query") == "true" {
		signedURL = h.signer.SignedURLQuery(link)
	} else {
		signedURL = h.signer.SignedURL(link)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"signed_url": signedURL,
	}); err != nil {
		log.Printf("share link generation failed: response encode error for %s: %v", r.URL.Path, err)
	}
}
