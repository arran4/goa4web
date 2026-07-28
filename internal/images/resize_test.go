package images

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"testing"
)

func TestParseDimension(t *testing.T) {
	tests := []struct {
		name    string
		dimStr  string
		wantW   int
		wantH   int
		wantErr bool
	}{
		{"Valid dimension", "1024x768", 1024, 768, false},
		{"Invalid format (no x)", "1024-768", 0, 0, true},
		{"Invalid format (multiple x)", "1024x768x2", 0, 0, true},
		{"Invalid width", "abcx768", 0, 0, true},
		{"Invalid height", "1024xdef", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH, err := ParseDimension(tt.dimStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDimension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotW != tt.wantW {
					t.Errorf("ParseDimension() gotW = %v, want %v", gotW, tt.wantW)
				}
				if gotH != tt.wantH {
					t.Errorf("ParseDimension() gotH = %v, want %v", gotH, tt.wantH)
				}
			}
		})
	}
}

func TestGenerateSafeSize(t *testing.T) {
	createImage := func(w, h int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		return img
	}

	t.Run("Invalid max dimensions", func(t *testing.T) {
		src := createImage(100, 100)
		_, err := GenerateSafeSize(src, ".jpg", "bild", 0, 100)
		if err == nil {
			t.Error("Expected error for maxWidth <= 0")
		}
		_, err = GenerateSafeSize(src, ".jpg", "bild", 100, -1)
		if err == nil {
			t.Error("Expected error for maxHeight <= 0")
		}
	})

	t.Run("Empty source image", func(t *testing.T) {
		src := image.NewRGBA(image.Rect(0, 0, 0, 0))
		_, err := GenerateSafeSize(src, ".jpg", "bild", 100, 100)
		if err == nil {
			t.Error("Expected error for empty source image")
		}
	})

	t.Run("Image already within safe size", func(t *testing.T) {
		src := createImage(80, 80)
		data, err := GenerateSafeSize(src, ".png", "bild", 100, 100)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		img, format, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to decode generated image: %v", err)
		}
		if format != "png" {
			t.Errorf("Expected png format, got %s", format)
		}
		if img.Bounds().Dx() != 80 || img.Bounds().Dy() != 80 {
			t.Errorf("Expected 80x80, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}
	})

	t.Run("Resize needed (width limiting)", func(t *testing.T) {
		src := createImage(200, 100)
		data, err := GenerateSafeSize(src, ".png", "bild", 100, 100)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to decode generated image: %v", err)
		}
		if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 50 {
			t.Errorf("Expected 100x50, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}
	})

	t.Run("Resize needed (height limiting)", func(t *testing.T) {
		src := createImage(100, 200)
		data, err := GenerateSafeSize(src, ".png", "bild", 100, 100)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to decode generated image: %v", err)
		}
		if img.Bounds().Dx() != 50 || img.Bounds().Dy() != 100 {
			t.Errorf("Expected 50x100, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}
	})

	t.Run("Resize to extreme (minimum 1x1)", func(t *testing.T) {
		src := createImage(1000, 1000)
		// Requesting 1x1 resize, but due to floating point and very small size it might be 1x1
		// Actually let's just make max width and max height very small compared to the original, e.g. 1
		// The math: ratio = Min(1/1000, 1/1000) = 0.001
		// newW = 1000 * 0.001 = 1, newH = 1000 * 0.001 = 1
		data, err := GenerateSafeSize(src, ".png", "bild", 1, 1)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to decode generated image: %v", err)
		}
		if img.Bounds().Dx() != 1 || img.Bounds().Dy() != 1 {
			t.Errorf("Expected 1x1, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}
	})

	t.Run("Generator: draw", func(t *testing.T) {
		src := createImage(200, 200)
		data, err := GenerateSafeSize(src, ".png", "draw", 100, 100)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to decode generated image: %v", err)
		}
		if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 100 {
			t.Errorf("Expected 100x100, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}
	})

	t.Run("Invalid extension", func(t *testing.T) {
		src := createImage(200, 200)
		_, err := GenerateSafeSize(src, ".invalid", "bild", 100, 100)
		if err == nil {
			t.Error("Expected error for invalid extension")
		}
	})

	t.Run("Invalid extension (no resize needed)", func(t *testing.T) {
		src := createImage(50, 50)
		_, err := GenerateSafeSize(src, ".invalid", "bild", 100, 100)
		if err == nil {
			t.Error("Expected error for invalid extension")
		}
	})
}
