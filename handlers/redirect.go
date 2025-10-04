package handlers

import "net/http"

// RedirectToGet redirects the client to targetURL. Non-idempotent methods such
// as POST are converted into a GET request using HTTP 303 to avoid resubmission.
// GET and HEAD requests continue to use HTTP 302 so their semantics remain
// unchanged.
func RedirectToGet(w http.ResponseWriter, r *http.Request, targetURL string) {
	http.Redirect(w, r, targetURL, redirectStatus(r))
}

func redirectStatus(r *http.Request) int {
	if r == nil {
		return http.StatusSeeOther
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		return http.StatusFound
	default:
		return http.StatusSeeOther
	}
}
