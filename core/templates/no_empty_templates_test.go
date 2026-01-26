package templates

import (
	"io/fs"
	"path/filepath"
	"testing"
)

func TestNoEmptyTemplates(t *testing.T) {
	// Walk the embeddedFS and ensure no files are empty.
	err := fs.WalkDir(embeddedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		// We are interested in templates.
		if ext != ".gohtml" && ext != ".gotxt" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() == 0 {
			t.Errorf("Template %s is empty", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir failed: %v", err)
	}
}
