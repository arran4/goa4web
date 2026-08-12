package user

import (
	"net/http"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func HasWebAuthn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if _, err := cd.GetWebAuthn(); err != nil {
			http.Error(w, "WebAuthn not configured", http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	}
}
