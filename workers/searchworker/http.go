package searchworker

import "net/http"

func redirectSeeOtherWithError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		http.Redirect(w, r, "", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "?error="+err.Error(), http.StatusSeeOther)
}
