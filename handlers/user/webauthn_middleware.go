package user

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
)

func HasWebAuthn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd.WebAuthn == nil {
			http.Error(w, "WebAuthn not configured", http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	}
}
