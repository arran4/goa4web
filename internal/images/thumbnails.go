package images

import (
	"bytes"
	"fmt"
	"image"

	"github.com/anthonynsimon/bild/transform"
	"golang.org/x/image/draw"
)

// ThumbnailGenerator represents a strategy for generating thumbnails.
type ThumbnailGenerator interface {
	Generate(srcImage image.Image, ext string, size int) ([]byte, error)
}

// thumbnailGenerators holds the registered thumbnail generator implementations.
var thumbnailGenerators = map[string]ThumbnailGenerator{
	"bild": &BildThumbnailGenerator{},
	"draw": &DrawThumbnailGenerator{},
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

// GenerateThumbnail creates a center-cropped square thumbnail from the source image using a specific generator and size.
func GenerateThumbnail(srcImage image.Image, ext string, generatorName string, size int) ([]byte, error) {
	return GetThumbnailGenerator(generatorName).Generate(srcImage, ext, size)
}

// GenerateThumbnailWithinBounds scales an image to fit the supplied height and width while preserving its aspect ratio.
func GenerateThumbnailWithinBounds(srcImage image.Image, ext string, generatorName string, maxHeight, maxWidth int) ([]byte, error) {
	return GenerateSafeSize(srcImage, ext, generatorName, maxWidth, maxHeight)
}

// DimensionsWithinBounds returns the aspect-ratio-preserving dimensions that fit within the supplied bounds.
func DimensionsWithinBounds(srcImage image.Image, maxHeight, maxWidth int) (height, width int, err error) {
	if maxHeight <= 0 || maxWidth <= 0 {
		return 0, 0, fmt.Errorf("thumbnail bounds must be greater than zero")
	}
	bounds := srcImage.Bounds()
	width = bounds.Dx()
	height = bounds.Dy()
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("invalid source dimensions: %dx%d", width, height)
	}
	if height <= maxHeight && width <= maxWidth {
		return height, width, nil
	}
	ratio := min(float64(maxHeight)/float64(height), float64(maxWidth)/float64(width))
	return max(1, int(float64(height)*ratio)), max(1, int(float64(width)*ratio)), nil
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

func (g *BildThumbnailGenerator) Generate(srcImage image.Image, ext string, size int) ([]byte, error) {
	crop := getCrop(srcImage)

	var croppedImg image.Image
	if sub, ok := srcImage.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		croppedImg = sub.SubImage(crop)
	} else {
		croppedImg = transform.Crop(srcImage, crop)
	}
	thumb := transform.Resize(croppedImg, size, size, transform.Linear)

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

func (g *DrawThumbnailGenerator) Generate(srcImage image.Image, ext string, size int) ([]byte, error) {
	crop := getCrop(srcImage)
	var tbuf bytes.Buffer
	thumb := image.NewRGBA(image.Rect(0, 0, size, size))
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
