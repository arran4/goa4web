package admin

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/internal/adminapi"
)

// AdminAPISigned returns a matcher verifying the admin API signature.
func AdminAPISigned() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		auth := r.Header.Get("Authorization")
		const prefix = "Goa4web "
		if !strings.HasPrefix(auth, prefix) {
			return false
		}
		parts := strings.SplitN(strings.TrimPrefix(auth, prefix), ":", 2)
		if len(parts) != 2 {
			return false
		}
		ts, sig := parts[0], parts[1]
		signer := adminapi.NewSigner(AdminAPISecret)
		return signer.Verify(r.Method, r.URL.Path, ts, sig)
	}
}
