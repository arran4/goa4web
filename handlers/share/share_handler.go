package share

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/sign/signutil"
)

type ShareHandler struct {
	signKey string
}

func NewShareHandler(signKey string) *ShareHandler {
	return &ShareHandler{signKey: signKey}
}

func (h *ShareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Query().Get("link")
	if link == "" {
		http.Error(w, "link parameter is required", http.StatusBadRequest)
		return
	}
	log.Printf("Generating share link (Handler) for: %s, UseQuery: %s", link, r.URL.Query().Get("use_query"))

	// Inject /shared/ into path
	sharedPath := signutil.InjectShared(link)

	// Generate nonce
	nonce := signutil.GenerateNonce()

	// Sign the URL
	var signedURL string
	var err error
	if r.URL.Query().Get("use_query") == "true" {
		signedURL, err = signutil.SignAndAddQuery(sharedPath, sharedPath, h.signKey, sign.WithNonce(nonce))
	} else {
		signedURL, err = signutil.SignAndAddPath(sharedPath, sharedPath, h.signKey, sign.WithNonce(nonce))
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
