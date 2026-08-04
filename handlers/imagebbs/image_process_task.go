package imagebbs

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"path"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload"
	"golang.org/x/image/draw"
)

// ProcessImageTask handles background thumbnail generation.
type ProcessImageTask struct {
	tasks.TaskString
	Config *config.RuntimeConfig
	ShaHex string
	Ext    string
}

var _ tasks.Task = (*ProcessImageTask)(nil)
var _ tasks.BackgroundTasker = (*ProcessImageTask)(nil)

func (t *ProcessImageTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if t.Config == nil {
		return nil, fmt.Errorf("config missing")
	}
	p := upload.ProviderFromConfig(t.Config)
	if p == nil {
		return nil, fmt.Errorf("provider missing")
	}

	sub1, sub2 := t.ShaHex[:2], t.ShaHex[2:4]
	fname := t.ShaHex + t.Ext

	data, err := p.Read(ctx, path.Join(sub1, sub2, fname))
	if err != nil {
		return nil, fmt.Errorf("read image fail %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image error %w", err)
	}

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
	thumb := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), img, crop, draw.Over, nil)

	var buf bytes.Buffer
	enc, err := imagesign.EncoderByExtension(t.Ext)
	if err != nil {
		return nil, fmt.Errorf("encoder fail %w", err)
	}
	if err := enc(&buf, thumb); err != nil {
		return nil, fmt.Errorf("encode thumb fail %w", err)
	}
	thumbName := t.ShaHex + "_thumb" + t.Ext
	if err := p.Write(ctx, path.Join(sub1, sub2, thumbName), buf.Bytes()); err != nil {
		return nil, fmt.Errorf("thumb write fail %w", err)
	}

	return nil, nil
}
