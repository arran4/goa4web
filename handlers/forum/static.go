package forum

import (
	_ "embed"
	"net/http"
)

//go:embed forum.js
var forumJS string

//go:embed forum.css
var forumCSS string

func (h *Handlers) serveJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(forumJS))
}

func (h *Handlers) serveCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Write([]byte(forumCSS))
}
