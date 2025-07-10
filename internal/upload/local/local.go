package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arran4/goa4web/internal/upload"
	"github.com/arran4/goa4web/runtimeconfig"
)

var (
	mkdirAll  = os.MkdirAll
	stat      = os.Stat
	writeFile = os.WriteFile
	readFile  = os.ReadFile
	remove    = os.Remove
	walkDir   = filepath.WalkDir
)

type Provider struct {
	Dir string
}

func providerFromConfig(cfg runtimeconfig.RuntimeConfig) upload.Provider {
	return Provider{Dir: cfg.ImageUploadDir}
}

func Register() { upload.RegisterProvider("local", providerFromConfig) }

func (p Provider) Check(ctx context.Context) error {
	if err := mkdirAll(p.Dir, 0o755); err != nil {
		return err
	}
	info, err := stat(p.Dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid dir")
	}
	test := filepath.Join(p.Dir, ".check")
	if err := writeFile(test, []byte("ok"), 0o644); err != nil {
		return fmt.Errorf("not writable")
	}
	remove(test)
	return nil
}

func (p Provider) Write(ctx context.Context, name string, data []byte) error {
	dir := filepath.Dir(filepath.Join(p.Dir, name))
	if err := mkdirAll(dir, 0o755); err != nil {
		return err
	}
	return writeFile(filepath.Join(p.Dir, name), data, 0o644)
}

func (p Provider) Read(ctx context.Context, name string) ([]byte, error) {
	return readFile(filepath.Join(p.Dir, name))
}

func (p Provider) Cleanup(ctx context.Context, limit int64) error {
	if limit <= 0 {
		return nil
	}
	type fileInfo struct {
		path string
		info os.FileInfo
	}
	var files []fileInfo
	var total int64
	walkDir(p.Dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if rel, err := filepath.Rel(p.Dir, path); err != nil || strings.HasPrefix(rel, "..") {
			return nil
		}
		files = append(files, fileInfo{path: path, info: info})
		total += info.Size()
		return nil
	})
	if total <= limit {
		return nil
	}
	sort.Slice(files, func(i, j int) bool { return files[i].info.ModTime().Before(files[j].info.ModTime()) })
	for _, f := range files {
		if total <= limit {
			break
		}
		if err := remove(f.path); err == nil {
			total -= f.info.Size()
		}
	}
	return nil
}
