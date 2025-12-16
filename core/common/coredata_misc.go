package common

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	"log"
	"path"

	"github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/upload"
	"golang.org/x/image/draw"
)

// StoreImageParams groups input for StoreImage.
type StoreImageParams struct {
	ID         string
	Ext        string
	Data       []byte
	Image      image.Image
	UploaderID int32
}

// StoreImage stores the image bytes, generates thumbnails and records metadata.
func (cd *CoreData) StoreImage(p StoreImageParams) (string, error) {
	if cd == nil || cd.queries == nil {
		return "", fmt.Errorf("no queries")
	}
	cfg := cd.Config
	sub1, sub2 := p.ID[:2], p.ID[2:4]
	fname := p.ID + p.Ext
	if prov := upload.ProviderFromConfig(cfg); prov != nil {
		if err := prov.Write(cd.ctx, path.Join(sub1, sub2, fname), p.Data); err != nil {
			log.Printf("upload write: %v", err)
			return "", fmt.Errorf("upload write %w", err)
		}
	}
	width := p.Image.Bounds().Dx()
	height := p.Image.Bounds().Dy()
	src := p.Image.Bounds()
	var crop image.Rectangle
	if src.Dx() > src.Dy() {
		side := src.Dy()
		x0 := src.Min.X + (src.Dx()-side)/2
		crop = image.Rect(x0, src.Min.Y, x0+side, src.Min.Y+side)
	} else {
		side := src.Dx()
		y0 := src.Min.Y + (src.Dy()-side)/2
		crop = image.Rect(src.Min.X, y0, src.Min.X+side, src.Min.Y+side)
	}
	thumbName := p.ID + "_thumb" + p.Ext
	var tbuf bytes.Buffer
	thumb := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), p.Image, crop, draw.Over, nil)
	enc, err := imagesign.EncoderByExtension(p.Ext)
	if err != nil {
		return "", fmt.Errorf("encoder ext %w", err)
	}
	if err := enc(&tbuf, thumb); err != nil {
		return "", fmt.Errorf("thumb encode %w", err)
	}
	if cp := upload.CacheProviderFromConfig(cfg); cp != nil {
		if err := cp.Write(cd.ctx, path.Join(sub1, sub2, thumbName), tbuf.Bytes()); err != nil {
			log.Printf("cache write: %v", err)
			return "", fmt.Errorf("cache write %w", err)
		}
		if ccp, ok := cp.(upload.CacheProvider); ok {
			if err := ccp.Cleanup(cd.ctx, int64(cfg.ImageCacheMaxBytes)); err != nil {
				log.Printf("cache cleanup: %v", err)
			}
		}
	}
	url := path.Join("/uploads", sub1, sub2, fname)
	_, err = cd.queries.CreateUploadedImageForUploader(cd.ctx, db.CreateUploadedImageForUploaderParams{
		UploaderID: p.UploaderID,
		Path:       sql.NullString{String: url, Valid: true},
		Width:      sql.NullInt32{Int32: int32(width), Valid: true},
		Height:     sql.NullInt32{Int32: int32(height), Valid: true},
		FileSize:   int32(len(p.Data)),
		GranteeID:  sql.NullInt32{Int32: p.UploaderID, Valid: p.UploaderID != 0},
	})
	if err != nil {
		return "", fmt.Errorf("create uploaded image %w", err)
	}
	return fname, nil
}
