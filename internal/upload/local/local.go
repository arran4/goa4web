package local

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/upload"
)

type FileSystem interface {
	MkdirAll(path string, perm fs.FileMode) error
	Stat(path string) (fs.FileInfo, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	ReadFile(path string) ([]byte, error)
	Remove(path string) error
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type osFS struct{}

func (osFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
func (osFS) Stat(path string) (fs.FileInfo, error)        { return os.Stat(path) }
func (osFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}
func (osFS) ReadFile(path string) ([]byte, error)         { return os.ReadFile(path) }
func (osFS) Remove(path string) error                     { return os.Remove(path) }
func (osFS) WalkDir(root string, fn fs.WalkDirFunc) error { return filepath.WalkDir(root, fn) }

type Provider struct {
	Dir string
	FS  FileSystem
}

func (p Provider) fs() FileSystem {
	if p.FS == nil {
		return osFS{}
	}
	return p.FS
}

// safePath verifies that name is a relative, non-traversing path and returns
// the path joined with the provider directory. fs.ValidPath implements the
// same rules as the standard library for opening files within an fs.FS.
func (p Provider) safePath(name string) (string, error) {
	if !fs.ValidPath(name) || filepath.IsAbs(name) {
		return "", fmt.Errorf("invalid path")
	}
	path := filepath.Join(p.Dir, filepath.Clean(name))
	if rel, err := filepath.Rel(p.Dir, path); err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid path")
	}
	return path, nil
}

func providerFromConfig(cfg config.RuntimeConfig) upload.Provider {
	return Provider{Dir: cfg.ImageUploadDir, FS: osFS{}}
}

func Register() { upload.RegisterProvider("local", providerFromConfig) }

func (p Provider) Check(ctx context.Context) error {
	fs := p.fs()
	if err := fs.MkdirAll(p.Dir, 0o755); err != nil {
		return err
	}
	info, err := fs.Stat(p.Dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid dir")
	}
	test := filepath.Join(p.Dir, ".check")
	if err := fs.WriteFile(test, []byte("ok"), 0o644); err != nil {
		return fmt.Errorf("not writable")
	}
	fs.Remove(test)
	return nil
}

func (p Provider) Write(ctx context.Context, name string, data []byte) error {
	fs := p.fs()
	path, err := p.safePath(name)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return fs.WriteFile(path, data, 0o644)
}

func (p Provider) Read(ctx context.Context, name string) ([]byte, error) {
	path, err := p.safePath(name)
	if err != nil {
		return nil, err
	}
	return p.fs().ReadFile(path)
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
	fs := p.fs()
	fs.WalkDir(p.Dir, func(path string, d os.DirEntry, err error) error {
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
		if err := fs.Remove(f.path); err == nil {
			total -= f.info.Size()
		}
	}
	return nil
}
