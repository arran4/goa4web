package share

import (
	"encoding/json"
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
	signer := sharesign.NewSigner(cd.Config, cd.Config.ShareSignSecret)
	signedURL := signer.SignedURL(link)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"signed_url": signedURL,
	})
}
