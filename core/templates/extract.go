package templates

import (
	"io/fs"
	"os"
	"path/filepath"
)

// WriteToDir writes the embedded templates and assets to dir preserving their structure.
func WriteToDir(dir string) error {
	for _, d := range []string{"site", "notifications", "email", "assets"} {
		if err := copyDir(d, dir); err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dstRoot string) error {
	fsys, err := fs.Sub(embeddedFS, src)
	if err != nil {
		return err
	}
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		out := filepath.Join(dstRoot, src, path)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
}
