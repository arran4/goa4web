package images

import (
	"bytes"
	"fmt"
	"image"
)

const (
	// maxDecodePixels caps the allocation for a server-side raster decode.
	maxDecodePixels = 40_000_000
	// maxDecodeSide bounds per-row and per-column allocations.
	maxDecodeSide = 16_384
)

// CheckDecodeDimensions rejects rasters that are too large for the server to decode.
func CheckDecodeDimensions(cfg image.Config) error {
	if cfg.Width <= 0 || cfg.Height <= 0 ||
		cfg.Width > maxDecodeSide || cfg.Height > maxDecodeSide ||
		cfg.Width > maxDecodePixels/cfg.Height {
		return fmt.Errorf("image dimensions exceed the server decode limit")
	}
	return nil
}

// DecodeForResize checks the dimensions before allocating decoded pixels.
func DecodeForResize(body []byte) (image.Image, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if err := CheckDecodeDimensions(cfg); err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(body))
	return img, err
}
