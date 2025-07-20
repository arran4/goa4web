package images

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/config"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload"
	"github.com/disintegration/imaging"
)

// UploadImageTask processes authenticated image uploads.
type UploadImageTask struct{ tasks.TaskString }

var uploadImageTask = &UploadImageTask{TaskString: TaskUploadImage}

func (UploadImageTask) Action(w http.ResponseWriter, r *http.Request) {
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

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	uid := int32(0)
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
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
