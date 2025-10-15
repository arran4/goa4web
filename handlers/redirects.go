package handlers

import (
	"net/http"
	"strings"
)

// RedirectSeeOther issues an HTTP 303 redirect to the provided target URL.
func RedirectSeeOther(w http.ResponseWriter, r *http.Request, target string) {
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// RedirectSeeOtherWithError appends an error query parameter to the target and issues an HTTP 303 redirect.
// When target is empty the redirect remains on the current path ("?error=").
func RedirectSeeOtherWithError(w http.ResponseWriter, r *http.Request, target string, err error) {
	if err == nil {
		RedirectSeeOther(w, r, target)
		return
	}
	RedirectSeeOtherWithMessage(w, r, target, err.Error())
}

// RedirectSeeOtherWithMessage appends a query parameter with the provided message and issues an HTTP 303 redirect.
func RedirectSeeOtherWithMessage(w http.ResponseWriter, r *http.Request, target, message string) {
	RedirectSeeOther(w, r, appendQuery(target, "error", message))
}

func appendQuery(target, key, value string) string {
	if value == "" {
		return target
	}
	if target == "" {
		return "?" + key + "=" + value
	}
	separator := "?"
	if strings.Contains(target, "?") {
		separator = "&"
	}
	return target + separator + key + "=" + value
}
