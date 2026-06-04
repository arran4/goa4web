package images

import (
	"bytes"
	"fmt"
	"image"

	"github.com/anthonynsimon/bild/transform"
)

var DefaultThumbSize = 200

// GenerateThumbnail creates a 200x200 center-cropped thumbnail from the source image.
func GenerateThumbnail(srcImage image.Image, ext string) ([]byte, error) {
	src := srcImage.Bounds()
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

	var croppedImg image.Image
	if sub, ok := srcImage.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		croppedImg = sub.SubImage(crop)
	} else {
		croppedImg = transform.Crop(srcImage, crop)
	}
	thumb := transform.Resize(croppedImg, DefaultThumbSize, DefaultThumbSize, transform.Linear)

	var tbuf bytes.Buffer
	enc, err := EncoderByExtension(ext)
	if err != nil {
		return nil, fmt.Errorf("encoder ext %w", err)
	}
	if err := enc(&tbuf, thumb); err != nil {
		return nil, fmt.Errorf("thumb encode %w", err)
	}
	return tbuf.Bytes(), nil
}
