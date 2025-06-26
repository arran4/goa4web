package goa4web

import (
	"net/http"
)

// taskRedirectWithoutQueryArgs redirects the request to the same URL path
// stripped of any query parameters using an HTTP 307 Temporary Redirect.
func taskRedirectWithoutQueryArgs(w http.ResponseWriter, r *http.Request) {
	u := r.URL
	u.RawQuery = ""
	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
}
