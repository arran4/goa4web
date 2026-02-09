package tasks

import "net/http"

// Page represents a page that can be rendered via GET request.
type Page interface {
	Get(w http.ResponseWriter, r *http.Request)
}
