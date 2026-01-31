package images

import (
	"bytes"
	"fmt"
	"image"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
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

			query := r.URL.Query()
			sig := query.Get("sig")
			tsStr := query.Get("ts")
			nonce := query.Get("nonce")

			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

			var err error
			var data string

			if tsStr != "" || nonce != "" {
				var opts []sign.SignOption
				if tsStr != "" {
					ts, _ := strconv.ParseInt(tsStr, 10, 64)
					opts = append(opts, sign.WithExpiry(time.Unix(ts, 0)))
				}
				if nonce != "" {
					opts = append(opts, sign.WithNonce(nonce))
				}
				// Verify "image:ID" or "cache:ID"
				data = prefix + id
				err = sign.Verify(data, sig, cd.ImageSignKey, opts...)
			} else {
				query.Del("ts")
				query.Del("sig")
				data = id
				if encoded := query.Encode(); encoded != "" {
					data = data + "?" + encoded
				}
				if prefix != "" {
					data = prefix + data
				}
				err = sign.Verify(data, sig, cd.ImageSignKey, sign.WithOutNonce())
			}

			if cd.ImageSignKey == "" || err != nil {
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
	if _, err := os.Stat(full); os.IsNotExist(err) {
		var opts []templates.Option
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, templates.WithDir(cfg.TemplatesDir))
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeContent(w, r, "missing_image.svg", time.Time{}, bytes.NewReader(templates.GetMissingImageData(opts...)))
		return
	}
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
