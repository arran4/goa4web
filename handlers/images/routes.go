package images

import (
	"bytes"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	handlers "github.com/arran4/goa4web/handlers"
	imagesign "github.com/arran4/goa4web/internal/images"
	router "github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/upload"
)

func verifyMiddleware(prefix string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			ts := r.URL.Query().Get("ts")
			sig := r.URL.Query().Get("sig")
			data := id
			if prefix != "" {
				data = prefix + id
			}
			if !imagesign.Verify(data, ts, sig) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RegisterRoutes attaches the image endpoints to r.
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/images/upload/image", tasks.Action(uploadImageTask)).
		Methods(http.MethodPost).
		MatcherFunc(handlers.RequiresAnAccount()).
		MatcherFunc(uploadImageTask.Matcher())
	r.HandleFunc("/images/pasteimg.js", handlers.PasteImageJS).Methods(http.MethodGet)
	r.Handle("/images/image/{id}", verifyMiddleware("image:")(http.HandlerFunc(serveImage))).Methods(http.MethodGet)
	r.Handle("/images/cache/{id}", verifyMiddleware("cache:")(http.HandlerFunc(serveCache))).Methods(http.MethodGet)
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if len(id) < 4 {
		http.NotFound(w, r)
		return
	}
	sub1, sub2 := id[:2], id[2:4]
	full := filepath.Join(config.AppRuntimeConfig.ImageUploadDir, sub1, sub2, id)
	http.ServeFile(w, r, full)
}

func serveCache(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if len(id) < 4 {
		http.NotFound(w, r)
		return
	}
	sub1, sub2 := id[:2], id[2:4]
	key := path.Join(sub1, sub2, id)
	if p := upload.CacheProviderFromConfig(config.AppRuntimeConfig); p != nil {
		data, err := p.Read(r.Context(), key)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeContent(w, r, id, time.Now(), bytes.NewReader(data))
		return
	}
	full := filepath.Join(config.AppRuntimeConfig.ImageCacheDir, sub1, sub2, id)
	http.ServeFile(w, r, full)
}

// Register registers the images router module.
func Register() {
	router.RegisterModule("images", nil, RegisterRoutes)
}
