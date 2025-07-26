package images

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload"
	"golang.org/x/image/draw"
)

// UploadImageTask processes authenticated image uploads.
type UploadImageTask struct{ tasks.TaskString }

var uploadImageTask = &UploadImageTask{TaskString: TaskUploadImage}

// ensure UploadImageTask conforms to tasks.Task
var _ tasks.Task = (*UploadImageTask)(nil)

func (UploadImageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config

	r.Body = http.MaxBytesReader(w, r.Body, int64(cfg.ImageMaxBytes))
	if err := r.ParseMultipartForm(int64(cfg.ImageMaxBytes)); err != nil {
		return fmt.Errorf("bad upload %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		return fmt.Errorf("image required %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read file %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	size := int64(len(data))

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("decode image %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	id := r.FormValue("id")
	if id == "" {
		id = fmt.Sprintf("%x", sha1.Sum(data))
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	sub1, sub2 := id[:2], id[2:4]
	fname := id + ext
	if p := upload.ProviderFromConfig(*cfg); p != nil {
		if err := p.Write(r.Context(), path.Join(sub1, sub2, fname), data); err != nil {
			log.Printf("upload write: %v", err)
			return fmt.Errorf("upload write %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	src := img.Bounds()
	var crop image.Rectangle
	if src.Dx() > src.Dy() {
		side := src.Dy()
		x0 := src.Min.X + (src.Dx()-side)/2
		crop = image.Rect(x0, src.Min.Y, x0+side, src.Min.Y+side)
	} else {
		side := src.Dx()
		y0 := src.Min.Y + (src.Dy()-side)/2
		crop = image.Rect(src.Min.X, y0, src.Min.X+side, y0+side)
	}
	thumbName := id + "_thumb" + ext
	var tbuf bytes.Buffer
	thumb := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), img, crop, draw.Over, nil)
	enc, err := imagesign.EncoderByExtension(ext)
	if err != nil {
		return fmt.Errorf("encoder ext %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := enc(&tbuf, thumb); err != nil {
		return fmt.Errorf("thumb encode %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cp := upload.CacheProviderFromConfig(*cfg); cp != nil {
		if err := cp.Write(r.Context(), path.Join(sub1, sub2, thumbName), tbuf.Bytes()); err != nil {
			log.Printf("cache write: %v", err)
			return fmt.Errorf("cache write %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if ccp, ok := cp.(upload.CacheProvider); ok {
			if err := ccp.Cleanup(r.Context(), int64(cfg.ImageCacheMaxBytes)); err != nil {
				log.Printf("cache cleanup: %v", err)
			}
		}
	}

	url := path.Join("/uploads", sub1, sub2, fname)

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
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
		return fmt.Errorf("create uploaded image %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd.ImageSigner != nil {
		signed := cd.ImageSigner.SignedRef("image:" + fname)
		return handlers.TextByteWriter([]byte(signed))
	}
	return handlers.TextByteWriter([]byte("image:" + fname))
}
