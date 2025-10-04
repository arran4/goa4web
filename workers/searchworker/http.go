package searchworker

import "net/http"

func redirectToGet(w http.ResponseWriter, r *http.Request, target string) {
	status := http.StatusSeeOther
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		status = http.StatusFound
	}
	http.Redirect(w, r, target, status)
}
