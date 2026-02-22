package images

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestEncoderByExtension(t *testing.T) {
	// Create a simple 1x1 image for testing
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})

	tests := []struct {
		ext        string
		wantFormat string
		wantErr    bool
	}{
		{ext: ".jpg", wantFormat: "jpeg"},
		{ext: ".JPG", wantFormat: "jpeg"},
		{ext: ".jpeg", wantFormat: "jpeg"},
		{ext: ".png", wantFormat: "png"},
		{ext: ".gif", wantFormat: "gif"},
		{ext: ".bmp", wantErr: true},
		{ext: ".txt", wantErr: true},
		{ext: "", wantErr: true},
		{ext: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			encodeFunc, err := EncoderByExtension(tt.ext)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncoderByExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if encodeFunc == nil {
				t.Fatal("EncoderByExtension() returned nil function")
			}

			var buf bytes.Buffer
			if err := encodeFunc(&buf, img); err != nil {
				t.Errorf("encoder function failed: %v", err)
				return
			}

			_, format, err := image.Decode(&buf)
			if err != nil {
				t.Errorf("image.Decode() failed: %v", err)
				return
			}

			if format != tt.wantFormat {
				t.Errorf("image.Decode() format = %v, want %v", format, tt.wantFormat)
			}
		})
	}
}
