package share

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
)

func RegisterShareRoutes(r *mux.Router, cfg *config.RuntimeConfig, shareSignKey string) {
	newsShareHandler := NewShareHandler(shareSignKey)
	r.Handle("/api/news/share", newsShareHandler).Methods("GET")

	blogsShareHandler := NewShareHandler(shareSignKey)
	r.Handle("/api/blogs/share", blogsShareHandler).Methods("GET")

	writingsShareHandler := NewShareHandler(shareSignKey)
	r.Handle("/api/writings/share", writingsShareHandler).Methods("GET")

	forumShareHandler := NewShareHandler(shareSignKey)
	r.Handle("/api/forum/share", forumShareHandler).Methods("GET")

	ogHandler := NewOGImageHandler(shareSignKey)
	r.Handle("/api/og-image", ogHandler).Methods("GET", "HEAD")
	r.Handle("/api/og-image/ts/{ts}/sign/{sign}", ogHandler).Methods("GET", "HEAD")
	r.Handle("/api/og-image/{data}/ts/{ts}/sign/{sign}", ogHandler).Methods("GET", "HEAD")
	r.Handle("/api/og-image/nonce/{nonce}/sign/{sign}", ogHandler).Methods("GET", "HEAD")
	r.Handle("/api/og-image/{data}/nonce/{nonce}/sign/{sign}", ogHandler).Methods("GET", "HEAD")
}
