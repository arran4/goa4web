package images

import (
	"bytes"
	"fmt"
	"image"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	intimages "github.com/arran4/goa4web/internal/images"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/upload"
)

func verifyMiddleware(prefix string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			if !intimages.ValidID(id) {
				w.WriteHeader(http.StatusForbidden)
				handlers.RenderErrorPage(w, r, fmt.Errorf("invalid id"))
				return
			}

			cleanURL, sig, opts, err := sign.ExtractQuerySig(r.URL.String())
			if err != nil || sig == "" {
				w.WriteHeader(http.StatusForbidden)
				handlers.RenderErrorPage(w, r, fmt.Errorf("missing or invalid signature"))
				return
			}

			u, err := url.Parse(cleanURL)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				handlers.RenderErrorPage(w, r, fmt.Errorf("invalid url"))
				return
			}

			data := prefix + id
			if q := u.Query().Encode(); q != "" {
				data += "?" + q
			}
			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
			if cd.ImageSignKey == "" || sign.Verify(data, sig, cd.ImageSignKey, opts...) != nil {
				w.WriteHeader(http.StatusForbidden)
				handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RegisterRoutes attaches the image endpoints to r.
func RegisterRoutes(r *mux.Router, cfg *config.RuntimeConfig, _ *nav.Registry) {
	ir := r.PathPrefix("/images").Subrouter()
	ir.Use(handlers.IndexMiddleware(CustomIndex))
	ir.HandleFunc("/upload/image", handlers.TaskHandler(uploadImageTask)).
		Methods(http.MethodPost).
		MatcherFunc(handlers.RequiresAnAccount()).
		MatcherFunc(uploadImageTask.Matcher())
	ir.HandleFunc("/pasteimg.js", handlers.PasteImageJS(cfg)).Methods(http.MethodGet)
	ir.Handle("/image/{id}", verifyMiddleware("image:")(http.HandlerFunc(serveImage))).
		Methods(http.MethodGet)
	ir.Handle("/cache/{id}", verifyMiddleware("cache:")(http.HandlerFunc(serveCache))).
		Methods(http.MethodGet)
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !intimages.ValidID(id) {
		w.WriteHeader(http.StatusForbidden)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid id"))
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config
	sub1, sub2 := id[:2], id[2:4]
	full := filepath.Join(cfg.ImageUploadDir, sub1, sub2, id)
	http.ServeFile(w, r, full)
}

func serveCache(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !intimages.ValidID(id) {
		w.WriteHeader(http.StatusForbidden)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid id"))
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config
	sub1, sub2 := id[:2], id[2:4]
	key := path.Join(sub1, sub2, id)
	if p := upload.CacheProviderFromConfig(cfg); p != nil {
		data, err := p.Read(r.Context(), key)
		if err != nil {
			// Try to regenerate
			if strings.Contains(id, "_thumb.") {
				origID := strings.Replace(id, "_thumb.", ".", 1)
				origKey := path.Join(sub1, sub2, origID)
				if up := upload.ProviderFromConfig(cfg); up != nil {
					origBytes, err := up.Read(r.Context(), origKey)
					if err == nil {
						img, _, err := image.Decode(bytes.NewReader(origBytes))
						if err == nil {
							ext := filepath.Ext(id)
							thumbBytes, err := intimages.GenerateThumbnail(img, ext)
							if err == nil {
								if err := p.Write(r.Context(), key, thumbBytes); err == nil {
									data = thumbBytes
								}
							}
						}
					}
				}
			}
		}
		if data == nil {
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
