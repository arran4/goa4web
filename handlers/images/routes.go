package images

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	router "github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/upload"
	imagesign "github.com/arran4/goa4web/pkg/images"
	"github.com/disintegration/imaging"
)

// SetSigningKey stores the key used for signing URLs.
func SetSigningKey(k string) { imagesign.SetSigningKey(k) }

// SignedURL maps an image identifier to a signed URL.
func SignedURL(id string) string { return imagesign.SignedURL(id) }

// SignedCacheURL maps a cache identifier to a signed URL.
func SignedCacheURL(id string) string { return imagesign.SignedCacheURL(id) }

// MapURL converts image: or cache: references to fully signed HTTP URLs.
// Only img tags are transformed; other tags are left unchanged.
func MapURL(tag, val string) string { return imagesign.MapURL(tag, val) }

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
			if !verify(data, ts, sig) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func verify(data, tsStr, sig string) bool { return imagesign.Verify(data, tsStr, sig) }

// RegisterRoutes attaches the image endpoints to r.
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/images/upload/image", uploadHandler).Methods(http.MethodPost)
	r.HandleFunc("/images/pasteimg.js", hcommon.PasteImageJS).Methods(http.MethodGet)
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

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.AppRuntimeConfig.ImageMaxBytes))
	if err := r.ParseMultipartForm(int64(config.AppRuntimeConfig.ImageMaxBytes)); err != nil {
		http.Error(w, "bad upload", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	size := int64(len(data))

	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		http.Error(w, "invalid image", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		id = fmt.Sprintf("%x", sha1.Sum(data))
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	sub1, sub2 := id[:2], id[2:4]
	fname := id + ext
	if p := upload.ProviderFromConfig(config.AppRuntimeConfig); p != nil {
		if err := p.Write(r.Context(), path.Join(sub1, sub2, fname), data); err != nil {
			log.Printf("upload write: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	thumb := imaging.Thumbnail(img, 200, 200, imaging.Lanczos)
	thumbName := id + "_thumb" + ext
	var tbuf bytes.Buffer
	imgFmt, _ := imaging.FormatFromExtension(ext)
	if err := imaging.Encode(&tbuf, thumb, imgFmt); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if cp := upload.CacheProviderFromConfig(config.AppRuntimeConfig); cp != nil {
		if err := cp.Write(r.Context(), path.Join(sub1, sub2, thumbName), tbuf.Bytes()); err != nil {
			log.Printf("cache write: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if ccp, ok := cp.(upload.CacheProvider); ok {
			_ = ccp.Cleanup(r.Context(), int64(config.AppRuntimeConfig.ImageCacheMaxBytes))
		}
	}

	url := path.Join("/uploads", sub1, sub2, fname)

	queries := r.Context().Value(corecorecommon.KeyQueries).(*db.Queries)
	uid := int32(0)
	if cd, ok := r.Context().Value(corecorecommon.KeyCoreData).(*corecorecommon.CoreData); ok && cd != nil {
		uid = cd.UserID
	}
	_, err = queries.CreateUploadedImage(r.Context(), db.CreateUploadedImageParams{
		UsersIdusers: uid,
		Path:         sql.NullString{String: url, Valid: true},
		Width:        sql.NullInt32{Int32: int32(width), Valid: true},
		Height:       sql.NullInt32{Int32: int32(height), Valid: true},
		FileSize:     int32(size),
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(imagesign.SignedRef("image:" + fname)))
}

// Register registers the images router module.
func Register() {
	router.RegisterModule("images", nil, RegisterRoutes)
}
