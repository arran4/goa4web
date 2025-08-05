package images

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/upload"
)

// validID reports whether s consists solely of alphanumeric characters.
func validID(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return false
		}
	}
	return true
}

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
			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
			if cd.ImageSigner == nil || !cd.ImageSigner.Verify(data, ts, sig) {
				w.WriteHeader(http.StatusForbidden)
				handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RegisterRoutes attaches the image endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, _ *nav.Registry) {
	ir := r.PathPrefix("/images").Subrouter()
	ir.Use(handlers.IndexMiddleware(CustomIndex))
	ir.HandleFunc("/upload/image", handlers.TaskHandler(uploadImageTask)).
		Methods(http.MethodPost).
		MatcherFunc(handlers.RequiresAnAccount()).
		MatcherFunc(uploadImageTask.Matcher())
	ir.HandleFunc("/pasteimg.js", handlers.PasteImageJS).Methods(http.MethodGet)
	ir.Handle("/image/{id}", verifyMiddleware("image:")(http.HandlerFunc(serveImage))).
		Methods(http.MethodGet)
	ir.Handle("/cache/{id}", verifyMiddleware("cache:")(http.HandlerFunc(serveCache))).
		Methods(http.MethodGet)
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config
	sub1, sub2 := id[:2], id[2:4]
	full := filepath.Join(cfg.ImageUploadDir, sub1, sub2, id)
	http.ServeFile(w, r, full)
}

func serveCache(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config
	sub1, sub2 := id[:2], id[2:4]
	key := path.Join(sub1, sub2, id)
	if p := upload.CacheProviderFromConfig(cfg); p != nil {
		data, err := p.Read(r.Context(), key)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeContent(w, r, id, time.Now(), bytes.NewReader(data))
		return
	}
	full := filepath.Join(cfg.ImageCacheDir, sub1, sub2, id)
	http.ServeFile(w, r, full)
}

// Register registers the images router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("images", nil, RegisterRoutes)
}
