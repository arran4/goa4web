package search

import (
	"net/http"
)

func SearchPage(w http.ResponseWriter, r *http.Request) {
	t := NewSearchTask().(*searchTask)
	t.Get(w, r)
}
