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
		{
			name:    "valid normal size",
			dimStr:  "1024x768",
			wantW:   1024,
			wantH:   768,
			wantErr: false,
		},
		{
			name:    "valid small size",
			dimStr:  "1x1",
			wantW:   1,
			wantH:   1,
			wantErr: false,
		},
		{
			name:    "invalid with spaces",
			dimStr:  " 1024 x 768 ",
			wantErr: true,
		},
		{
			name:    "invalid missing x",
			dimStr:  "1024",
			wantErr: true,
		},
		{
			name:    "invalid empty string",
			dimStr:  "",
			wantErr: true,
		},
		{
			name:    "invalid multiple x",
			dimStr:  "1024x768x24",
			wantErr: true,
		},
		{
			name:    "invalid width not a number",
			dimStr:  "abcx768",
			wantErr: true,
		},
		{
			name:    "invalid height not a number",
			dimStr:  "1024xabc",
			wantErr: true,
		},
		{
			name:    "invalid missing height",
			dimStr:  "1024x",
			wantErr: true,
		},
		{
			name:    "invalid missing width",
			dimStr:  "x768",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h, err := ParseDimension(tt.dimStr)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ParseDimension(%q) error = %v, wantErr %v",
					tt.dimStr,
					err,
					tt.wantErr,
				)
				return
			}

			if tt.wantErr {
				return
			}

			if w != tt.wantW {
				t.Errorf("ParseDimension(%q) width = %d, want %d", tt.dimStr, w, tt.wantW)
			}
			if h != tt.wantH {
				t.Errorf("ParseDimension(%q) height = %d, want %d", tt.dimStr, h, tt.wantH)
			}
		})
	}
}

func TestGenerateSafeSize(t *testing.T) {
	createImage := func(w, h int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.Draw(
			img,
			img.Bounds(),
			&image.Uniform{C: color.White},
			image.Point{},
			draw.Src,
		)
		return img
	}

	t.Run("invalid max dimensions", func(t *testing.T) {
		src := createImage(100, 100)

		_, err := GenerateSafeSize(src, ".jpg", "bild", 0, 100)
		if err == nil {
			t.Error("expected error for maxWidth <= 0")
		}

		_, err = GenerateSafeSize(src, ".jpg", "bild", 100, -1)
		if err == nil {
			t.Error("expected error for maxHeight <= 0")
		}
	})

	t.Run("empty source image", func(t *testing.T) {
		src := image.NewRGBA(image.Rect(0, 0, 0, 0))

		_, err := GenerateSafeSize(src, ".jpg", "bild", 100, 100)
		if err == nil {
			t.Error("expected error for empty source image")
		}
	})

	t.Run("image already within safe size", func(t *testing.T) {
		src := createImage(80, 80)

		data, err := GenerateSafeSize(src, ".png", "bild", 100, 100)
		if err != nil {
			t.Fatalf("GenerateSafeSize() error = %v", err)
		}

		img, format, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("image.Decode() error = %v", err)
		}

		if format != "png" {
			t.Errorf("format = %q, want %q", format, "png")
		}
		if gotW, gotH := img.Bounds().Dx(), img.Bounds().Dy(); gotW != 80 || gotH != 80 {
			t.Errorf("dimensions = %dx%d, want 80x80", gotW, gotH)
		}
	})

	t.Run("resize needed with width limiting", func(t *testing.T) {
		src := createImage(200, 100)

		data, err := GenerateSafeSize(src, ".png", "bild", 100, 100)
		if err != nil {
			t.Fatalf("GenerateSafeSize() error = %v", err)
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("image.Decode() error = %v", err)
		}

		if gotW, gotH := img.Bounds().Dx(), img.Bounds().Dy(); gotW != 100 || gotH != 50 {
			t.Errorf("dimensions = %dx%d, want 100x50", gotW, gotH)
		}
	})

	t.Run("resize needed with height limiting", func(t *testing.T) {
		src := createImage(100, 200)

		data, err := GenerateSafeSize(src, ".png", "bild", 100, 100)
		if err != nil {
			t.Fatalf("GenerateSafeSize() error = %v", err)
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("image.Decode() error = %v", err)
		}

		if gotW, gotH := img.Bounds().Dx(), img.Bounds().Dy(); gotW != 50 || gotH != 100 {
			t.Errorf("dimensions = %dx%d, want 50x100", gotW, gotH)
		}
	})

	t.Run("resize to minimum one by one", func(t *testing.T) {
		src := createImage(1000, 1000)

		data, err := GenerateSafeSize(src, ".png", "bild", 1, 1)
		if err != nil {
			t.Fatalf("GenerateSafeSize() error = %v", err)
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("image.Decode() error = %v", err)
		}

		if gotW, gotH := img.Bounds().Dx(), img.Bounds().Dy(); gotW != 1 || gotH != 1 {
			t.Errorf("dimensions = %dx%d, want 1x1", gotW, gotH)
		}
	})

	t.Run("draw generator", func(t *testing.T) {
		src := createImage(200, 200)

		data, err := GenerateSafeSize(src, ".png", "draw", 100, 100)
		if err != nil {
			t.Fatalf("GenerateSafeSize() error = %v", err)
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("image.Decode() error = %v", err)
		}

		if gotW, gotH := img.Bounds().Dx(), img.Bounds().Dy(); gotW != 100 || gotH != 100 {
			t.Errorf("dimensions = %dx%d, want 100x100", gotW, gotH)
		}
	})

	t.Run("invalid extension", func(t *testing.T) {
		src := createImage(200, 200)

		_, err := GenerateSafeSize(src, ".invalid", "bild", 100, 100)
		if err == nil {
			t.Error("expected error for invalid extension")
		}
	})

	t.Run("invalid extension without resize", func(t *testing.T) {
		src := createImage(50, 50)

		_, err := GenerateSafeSize(src, ".invalid", "bild", 100, 100)
		if err == nil {
			t.Error("expected error for invalid extension")
		}
	})
}
