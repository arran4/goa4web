package images

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/transform"
	"golang.org/x/image/draw"
)

// ParseDimension parses a string like "1024x768" into width and height.
func ParseDimension(dimStr string) (int, int, error) {
	parts := strings.Split(dimStr, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid dimension format: %s", dimStr)
	}
	w, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width: %w", err)
	}
	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height: %w", err)
	}
	return w, h, nil
}

// GenerateSafeSize creates a resized version of an image if it exceeds the maximum dimensions.
// It preserves the aspect ratio.
func GenerateSafeSize(srcImage image.Image, ext string, generatorName string, maxWidth, maxHeight int) ([]byte, error) {
	if maxWidth <= 0 || maxHeight <= 0 {
		return nil, fmt.Errorf("max dimensions must be greater than zero")
	}

	bounds := srcImage.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid source image dimensions: %dx%d", w, h)
	}

	if w <= maxWidth && h <= maxHeight {
		// Image is already safe size, return original encoded
		var buf bytes.Buffer
		enc, err := EncoderByExtension(ext)
		if err != nil {
			return nil, fmt.Errorf("encoder ext %w", err)
		}
		if err := enc(&buf, srcImage); err != nil {
			return nil, fmt.Errorf("thumb encode %w", err)
		}
		return buf.Bytes(), nil
	}

	// Calculate new dimensions preserving aspect ratio
	ratio := math.Min(float64(maxWidth)/float64(w), float64(maxHeight)/float64(h))
	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	if generatorName == "draw" {
		thumb := image.NewRGBA(image.Rect(0, 0, newW, newH))
		draw.CatmullRom.Scale(thumb, thumb.Bounds(), srcImage, bounds, draw.Over, nil)

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

	// Default to bild
	thumb := transform.Resize(srcImage, newW, newH, transform.Linear)
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
