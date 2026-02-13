package routes

import (
	"net/http"
)

func BookmarksPage(w http.ResponseWriter, r *http.Request) {
	t := NewBookmarksTask().(*bookmarksTask)
	t.Get(w, r)
}
