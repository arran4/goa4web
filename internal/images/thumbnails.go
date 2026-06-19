package images

import (
	"bytes"
	"fmt"
	"image"

	"github.com/anthonynsimon/bild/transform"
	"golang.org/x/image/draw"
)

var DefaultThumbSize = 200

// ThumbnailGenerator represents a strategy for generating thumbnails.
type ThumbnailGenerator interface {
	Generate(srcImage image.Image, ext string) ([]byte, error)
}

// thumbnailGenerators holds the registered thumbnail generator implementations.
var thumbnailGenerators = map[string]ThumbnailGenerator{
	"bild":   &BildThumbnailGenerator{},
	"draw":   &DrawThumbnailGenerator{},
}

// GetThumbnailGenerator returns the thumbnail generator by name, defaulting to "bild".
func GetThumbnailGenerator(name string) ThumbnailGenerator {
	if g, ok := thumbnailGenerators[name]; ok {
		return g
	}
	return thumbnailGenerators["bild"]
}

// RegisterThumbnailGenerator registers a new thumbnail generator.
func RegisterThumbnailGenerator(name string, g ThumbnailGenerator) {
	thumbnailGenerators[name] = g
}

// BildThumbnailGenerator uses the bild transform library.
type BildThumbnailGenerator struct{}

// DrawThumbnailGenerator uses the standard golang.org/x/image/draw library.
type DrawThumbnailGenerator struct{}

// GenerateThumbnail creates a 200x200 center-cropped thumbnail from the source image using a specific generator.
func GenerateThumbnail(srcImage image.Image, ext string, generatorName string) ([]byte, error) {
	return GetThumbnailGenerator(generatorName).Generate(srcImage, ext)
}

func getCrop(srcImage image.Image) image.Rectangle {
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
	return crop
}

func (g *BildThumbnailGenerator) Generate(srcImage image.Image, ext string) ([]byte, error) {
	crop := getCrop(srcImage)

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

func (g *DrawThumbnailGenerator) Generate(srcImage image.Image, ext string) ([]byte, error) {
	crop := getCrop(srcImage)
	var tbuf bytes.Buffer
	thumb := image.NewRGBA(image.Rect(0, 0, DefaultThumbSize, DefaultThumbSize))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), srcImage, crop, draw.Over, nil)
	enc, err := EncoderByExtension(ext)
	if err != nil {
		return nil, fmt.Errorf("encoder ext %w", err)
	}
	if err := enc(&tbuf, thumb); err != nil {
		return nil, fmt.Errorf("thumb encode %w", err)
	}
	return tbuf.Bytes(), nil
}
