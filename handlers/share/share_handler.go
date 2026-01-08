package share

import (
	"encoding/json"
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

	signedURL := h.signer.SignedURL(link)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"signed_url": signedURL,
	})
}
