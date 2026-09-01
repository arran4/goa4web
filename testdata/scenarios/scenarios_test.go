package scenarios

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommittedImageAssetsDecode(t *testing.T) {
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	foundImages := 0
	err := fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !imageExts[ext] {
			return nil
		}

		foundImages++
		t.Run(path, func(t *testing.T) {
			file, err := FS.Open(path)
			if err != nil {
				t.Fatalf("Open %s: %v", path, err)
			}
			defer file.Close()

			img, format, err := image.Decode(file)
			if err != nil {
				t.Fatalf("image.Decode failed for %s: %v (format: %q)", path, err, format)
			}
			bounds := img.Bounds()
			if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
				t.Errorf("invalid image bounds for %s: %v", path, bounds)
			}
		})
		return nil
	})

	if err != nil {
		t.Fatalf("WalkDir failed: %v", err)
	}
	if foundImages == 0 {
		t.Fatal("expected to find committed image assets in scenarios, but found none")
	}
}
