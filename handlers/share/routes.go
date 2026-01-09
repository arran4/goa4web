package share

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sharesign"
)

func RegisterShareRoutes(r *mux.Router, cfg *config.RuntimeConfig, signer *sharesign.Signer) {
	newsShareHandler := NewShareHandler(signer)
	r.Handle("/api/news/share", newsShareHandler).Methods("GET")

	blogsShareHandler := NewShareHandler(signer)
	r.Handle("/api/blogs/share", blogsShareHandler).Methods("GET")

	writingsShareHandler := NewShareHandler(signer)
	r.Handle("/api/writings/share", writingsShareHandler).Methods("GET")

	forumShareHandler := NewShareHandler(signer)
	r.Handle("/api/forum/share", forumShareHandler).Methods("GET")

	r.HandleFunc("/api/og-image", OGImageHandler).Methods("GET")
}
