package images

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"testing"
)

func TestGenerateThumbnail(t *testing.T) {
	// Helper to create a solid color image
	createImage := func(w, h int, c color.Color) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
		return img
	}

	// 1. Test Dimensions and Formats
	t.Run("DimensionsAndFormats", func(t *testing.T) {
		src := createImage(400, 300, color.RGBA{255, 0, 0, 255})
		formats := []string{".jpg", ".jpeg", ".png", ".gif"}

		for _, ext := range formats {
			t.Run(ext, func(t *testing.T) {
				thumbData, err := GenerateThumbnail(src, ext, "bild", 200)
				if err != nil {
					t.Fatalf("GenerateThumbnail failed for %s: %v", ext, err)
				}

				thumb, format, err := image.Decode(bytes.NewReader(thumbData))
				if err != nil {
					t.Fatalf("Failed to decode generated thumbnail for %s: %v", ext, err)
				}

				if thumb.Bounds().Dx() != 200 || thumb.Bounds().Dy() != 200 {
					t.Errorf("Thumbnail dimensions = %dx%d, want 200x200", thumb.Bounds().Dx(), thumb.Bounds().Dy())
				}

				expectedFormat := "jpeg"
				if ext == ".png" {
					expectedFormat = "png"
				}
				if ext == ".gif" {
					expectedFormat = "gif"
				}

				if format != expectedFormat {
					t.Errorf("Thumbnail format = %s, want %s", format, expectedFormat)
				}
			})
		}
	})

	// 2. Test Invalid Extension
	t.Run("InvalidExtension", func(t *testing.T) {
		src := createImage(100, 100, color.Black)
		_, err := GenerateThumbnail(src, ".bmp", "bild", 200)
		if err == nil {
			t.Error("GenerateThumbnail should fail for .bmp")
		}
	})

	// 3. Test Cropping Logic (Landscape)
	t.Run("LandscapeCrop", func(t *testing.T) {
		// 300x100 image: Left(Red), Middle(Green), Right(Blue)
		// Each section is 100x100
		src := image.NewRGBA(image.Rect(0, 0, 300, 100))
		red := color.RGBA{255, 0, 0, 255}
		green := color.RGBA{0, 255, 0, 255}
		blue := color.RGBA{0, 0, 255, 255}

		draw.Draw(src, image.Rect(0, 0, 100, 100), &image.Uniform{red}, image.Point{}, draw.Src)
		draw.Draw(src, image.Rect(100, 0, 200, 100), &image.Uniform{green}, image.Point{}, draw.Src)
		draw.Draw(src, image.Rect(200, 0, 300, 100), &image.Uniform{blue}, image.Point{}, draw.Src)

		thumbData, err := GenerateThumbnail(src, ".png", "bild", 200)
		if err != nil {
			t.Fatalf("GenerateThumbnail failed: %v", err)
		}

		thumb, _, err := image.Decode(bytes.NewReader(thumbData))
		if err != nil {
			t.Fatalf("Failed to decode: %v", err)
		}

		// Check center pixel color
		// The thumbnail is 200x200. The source crop was the middle 100x100 (Green).
		// So the entire thumbnail should be Green.
		c := thumb.At(100, 100)
		r, g, b, _ := c.RGBA()

		// RGBA returns values in [0, 65535]. Green is (0, 65535, 0).
		if r > 1000 || g < 60000 || b > 1000 {
			t.Errorf("Center pixel color = (%d, %d, %d), want Green", r, g, b)
		}
	})

	// 4. Test Cropping Logic (Portrait)
	t.Run("PortraitCrop", func(t *testing.T) {
		// 100x300 image: Top(Red), Middle(Green), Bottom(Blue)
		// Each section is 100x100
		src := image.NewRGBA(image.Rect(0, 0, 100, 300))
		red := color.RGBA{255, 0, 0, 255}
		green := color.RGBA{0, 255, 0, 255}
		blue := color.RGBA{0, 0, 255, 255}

		draw.Draw(src, image.Rect(0, 0, 100, 100), &image.Uniform{red}, image.Point{}, draw.Src)
		draw.Draw(src, image.Rect(0, 100, 100, 200), &image.Uniform{green}, image.Point{}, draw.Src)
		draw.Draw(src, image.Rect(0, 200, 100, 300), &image.Uniform{blue}, image.Point{}, draw.Src)

		thumbData, err := GenerateThumbnail(src, ".png", "bild", 200)
		if err != nil {
			t.Fatalf("GenerateThumbnail failed: %v", err)
		}

		thumb, _, err := image.Decode(bytes.NewReader(thumbData))
		if err != nil {
			t.Fatalf("Failed to decode: %v", err)
		}

		// Check center pixel color
		c := thumb.At(100, 100)
		r, g, b, _ := c.RGBA()

		if r > 1000 || g < 60000 || b > 1000 {
			t.Errorf("Center pixel color = (%d, %d, %d), want Green", r, g, b)
		}
	})
}

func TestGenerateThumbnailWithinBoundsPreservesAspectRatio(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1200, 400))
	thumbData, err := GenerateThumbnailWithinBounds(src, ".png", "bild", 400, 800)
	if err != nil {
		t.Fatalf("GenerateThumbnailWithinBounds: %v", err)
	}
	thumb, _, err := image.Decode(bytes.NewReader(thumbData))
	if err != nil {
		t.Fatalf("decode thumbnail: %v", err)
	}
	if got, want := thumb.Bounds().Dx(), 800; got != want {
		t.Fatalf("thumbnail width = %d, want %d", got, want)
	}
	if got, want := thumb.Bounds().Dy(), 266; got != want {
		t.Fatalf("thumbnail height = %d, want %d", got, want)
	}

	height, width, err := DimensionsWithinBounds(src, 400, 800)
	if err != nil {
		t.Fatalf("DimensionsWithinBounds: %v", err)
	}
	if height != 266 || width != 800 {
		t.Fatalf("thumbnail dimensions = %dx%d, want 266x800", height, width)
	}
}

func TestDimensionsWithinBoundsEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		srcWidth  int
		srcHeight int
		maxWidth  int
		maxHeight int
		wantW     int
		wantH     int
		wantErr   bool
	}{
		{
			name:      "no scaling needed",
			srcWidth:  100,
			srcHeight: 100,
			maxWidth:  200,
			maxHeight: 200,
			wantW:     100,
			wantH:     100,
			wantErr:   false,
		},
		{
			name:      "exact match",
			srcWidth:  200,
			srcHeight: 200,
			maxWidth:  200,
			maxHeight: 200,
			wantW:     200,
			wantH:     200,
			wantErr:   false,
		},
		{
			name:      "constrained by width",
			srcWidth:  400,
			srcHeight: 200,
			maxWidth:  200,
			maxHeight: 200,
			wantW:     200,
			wantH:     100,
			wantErr:   false,
		},
		{
			name:      "constrained by height",
			srcWidth:  200,
			srcHeight: 400,
			maxWidth:  200,
			maxHeight: 200,
			wantW:     100,
			wantH:     200,
			wantErr:   false,
		},
		{
			name:      "extreme scaling down to minimum 1",
			srcWidth:  1000,
			srcHeight: 1,
			maxWidth:  2,
			maxHeight: 2,
			wantW:     2,
			wantH:     1,
			wantErr:   false,
		},
		{
			name:      "error max height 0",
			srcWidth:  100,
			srcHeight: 100,
			maxWidth:  100,
			maxHeight: 0,
			wantW:     0,
			wantH:     0,
			wantErr:   true,
		},
		{
			name:      "error invalid source width",
			srcWidth:  0,
			srcHeight: 100,
			maxWidth:  200,
			maxHeight: 200,
			wantW:     0,
			wantH:     0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := image.NewRGBA(image.Rect(0, 0, tt.srcWidth, tt.srcHeight))
			gotH, gotW, err := DimensionsWithinBounds(src, tt.maxHeight, tt.maxWidth)

			if (err != nil) != tt.wantErr {
				t.Errorf("DimensionsWithinBounds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotW != tt.wantW {
				t.Errorf("DimensionsWithinBounds() gotW = %v, want %v", gotW, tt.wantW)
			}

			if gotH != tt.wantH {
				t.Errorf("DimensionsWithinBounds() gotH = %v, want %v", gotH, tt.wantH)
			}
		})
	}
}
