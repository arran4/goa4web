package images

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gorilla/mux"

	corecommon "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	router "github.com/arran4/goa4web/internal/router"
	handlerspkg "github.com/arran4/goa4web/pkg/handlers"
	imagesign "github.com/arran4/goa4web/pkg/images"
	"github.com/arran4/goa4web/runtimeconfig"
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
	r.HandleFunc("/images/pasteimg.js", handlerspkg.PasteImageJS).Methods(http.MethodGet)
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
	full := filepath.Join(runtimeconfig.AppRuntimeConfig.ImageUploadDir, sub1, sub2, id)
	http.ServeFile(w, r, full)
}

func serveCache(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if len(id) < 4 {
		http.NotFound(w, r)
		return
	}
	sub1, sub2 := id[:2], id[2:4]
	full := filepath.Join(runtimeconfig.AppRuntimeConfig.ImageCacheDir, sub1, sub2, id)
	http.ServeFile(w, r, full)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(runtimeconfig.AppRuntimeConfig.ImageMaxBytes))
	if err := r.ParseMultipartForm(int64(runtimeconfig.AppRuntimeConfig.ImageMaxBytes)); err != nil {
		http.Error(w, "bad upload", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	info, err := os.Stat(runtimeconfig.AppRuntimeConfig.ImageUploadDir)
	if err != nil || !info.IsDir() {
		http.Error(w, "Uploads disabled", http.StatusInternalServerError)
		return
	}

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
	destDir := filepath.Join(runtimeconfig.AppRuntimeConfig.ImageUploadDir, sub1, sub2)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fname := id + ext
	fullPath := filepath.Join(destDir, fname)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	thumb := imaging.Thumbnail(img, 200, 200, imaging.Lanczos)
	cacheDir := filepath.Join(runtimeconfig.AppRuntimeConfig.ImageCacheDir, sub1, sub2)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	thumbName := id + "_thumb" + ext
	thumbPath := filepath.Join(cacheDir, thumbName)
	if err := imaging.Save(thumb, thumbPath); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	PruneCache(runtimeconfig.AppRuntimeConfig.ImageCacheDir, int64(runtimeconfig.AppRuntimeConfig.ImageCacheMaxBytes), 0)

	url := path.Join("/uploads", sub1, sub2, fname)

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	uid := int32(0)
	if cd, ok := r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData); ok && cd != nil {
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

type fileInfo struct {
	path string
	info os.FileInfo
}

// PruneCache removes old files until the total size is under limit.
func PruneCache(dir string, limit int64, verbosity int) {
	if limit <= 0 {
		return
	}
	var files []fileInfo
	var total int64
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if rel, err := filepath.Rel(dir, path); err != nil || strings.HasPrefix(rel, "..") {
			return nil
		}
		files = append(files, fileInfo{path, info})
		total += info.Size()
		return nil
	})
	if total <= limit {
		if verbosity > 1 {
			log.Printf("cache within limit: %d bytes", total)
		}
		return
	}
	if verbosity > 0 {
		log.Printf("pruning cache %s: %d bytes over limit", dir, total-limit)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].info.ModTime().Before(files[j].info.ModTime()) })
	for _, f := range files {
		if total <= limit {
			break
		}
		if err := os.Remove(f.path); err == nil {
			if verbosity > 1 {
				log.Printf("removed %s (%d bytes)", f.path, f.info.Size())
			}
			total -= f.info.Size()
		}
	}
}
