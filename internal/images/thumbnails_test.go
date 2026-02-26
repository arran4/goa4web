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
				thumbData, err := GenerateThumbnail(src, ext)
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
		_, err := GenerateThumbnail(src, ".bmp")
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

		thumbData, err := GenerateThumbnail(src, ".png")
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

		thumbData, err := GenerateThumbnail(src, ".png")
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
