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

	ogImageHandler := NewOGImageHandler(shareSignKey)
	ogImage := r.PathPrefix("/api/og-image/").Subrouter()
	ogImage.HandleFunc("/{data}/nonce/{nonce}/sign/{sign}", ogImageHandler.ServeHTTP)
	ogImage.HandleFunc("/{data}/ts/{ts}/sign/{sign}", ogImageHandler.ServeHTTP)
	ogImage.HandleFunc("/{data}/nonce/{nonce}", ogImageHandler.ServeHTTP)
	ogImage.HandleFunc("/{data}/ts/{ts}", ogImageHandler.ServeHTTP)
	ogImage.HandleFunc("/{data}", ogImageHandler.ServeHTTP)
}
