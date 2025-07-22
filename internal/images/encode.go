package images

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
)

// EncoderByExtension returns an encoder function for the given extension.
func EncoderByExtension(ext string) (func(io.Writer, image.Image) error, error) {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return func(w io.Writer, img image.Image) error { return jpeg.Encode(w, img, nil) }, nil
	case ".png":
		return func(w io.Writer, img image.Image) error { return png.Encode(w, img) }, nil
	case ".gif":
		return func(w io.Writer, img image.Image) error { return gif.Encode(w, img, nil) }, nil
	default:
		return nil, fmt.Errorf("unsupported image extension: %s", ext)
	}
}
