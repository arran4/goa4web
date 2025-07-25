package local

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
	"time"
)

type memFile struct {
	path string
	data []byte
	mod  time.Time
	dir  bool
}

func (f *memFile) Name() string { return filepath.Base(f.path) }
func (f *memFile) Size() int64  { return int64(len(f.data)) }
func (f *memFile) Mode() fs.FileMode {
	if f.dir {
		return fs.ModeDir
	}
	return 0
}
func (f *memFile) ModTime() time.Time { return f.mod }
func (f *memFile) IsDir() bool        { return f.dir }
func (f *memFile) Sys() any           { return nil }

type memDirEntry struct{ *memFile }

func (e memDirEntry) Type() fs.FileMode          { return e.Mode() }
func (e memDirEntry) Info() (fs.FileInfo, error) { return e.memFile, nil }

type memFS struct {
	files   map[string]*memFile
	counter int
}

func newMemFS() *memFS { return &memFS{files: map[string]*memFile{}} }

func (m *memFS) MkdirAll(path string, perm fs.FileMode) error {
	path = filepath.Clean(path)
	if path == "." || path == "/" {
		return nil
	}
	if _, ok := m.files[path]; !ok {
		m.files[path] = &memFile{path: path, dir: true, mod: time.Now()}
	}
	return nil
}

func (m *memFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	m.counter++
	_ = m.MkdirAll(filepath.Dir(path), perm)
	m.files[path] = &memFile{path: path, data: append([]byte(nil), data...), mod: time.Unix(int64(m.counter), 0)}
	return nil
}

func (m *memFS) ReadFile(path string) ([]byte, error) {
	f, ok := m.files[path]
	if !ok || f.dir {
		return nil, fs.ErrNotExist
	}
	return append([]byte(nil), f.data...), nil
}

func (m *memFS) Stat(path string) (fs.FileInfo, error) {
	f, ok := m.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return f, nil
}

func (m *memFS) Remove(path string) error {
	if _, ok := m.files[path]; !ok {
		return fs.ErrNotExist
	}
	delete(m.files, path)
	return nil
}

func (m *memFS) WalkDir(root string, fn fs.WalkDirFunc) error {
	for p, f := range m.files {
		if !f.dir {
			if err := fn(p, memDirEntry{f}, nil); err != nil && !errors.Is(err, fs.SkipDir) {
				return err
			}
		}
	}
	return nil
}

func TestCleanup(t *testing.T) {
	mfs := newMemFS()
	p := Provider{Dir: "/cache", FS: mfs}
	if err := p.Write(context.Background(), "a", []byte("1")); err != nil {
		t.Fatal(err)
	}
	if err := p.Write(context.Background(), "b", []byte("22")); err != nil {
		t.Fatal(err)
	}
	if err := p.Cleanup(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	if _, err := mfs.Stat("/cache/a"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a not removed: %v", err)
	}
	if _, err := mfs.Stat("/cache/b"); err != nil {
		t.Fatalf("b removed: %v", err)
	}
}

func TestWriteRejectsTraversal(t *testing.T) {
	mfs := newMemFS()
	p := Provider{Dir: "/cache", FS: mfs}
	if err := p.Write(context.Background(), "../evil", []byte("x")); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := mfs.Stat("/bad"); err == nil {
		t.Fatalf("file created")
	}
}

func TestWriteRejectsAbs(t *testing.T) {
	mfs := newMemFS()
	p := Provider{Dir: "/cache", FS: mfs}
	if err := p.Write(context.Background(), "/abs", []byte("x")); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadRejectsTraversal(t *testing.T) {
	mfs := newMemFS()
	_ = mfs.WriteFile("/cache/good", []byte("data"), 0o644)
	p := Provider{Dir: "/cache", FS: mfs}
	if _, err := p.Read(context.Background(), "../good"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadRejectsAbs(t *testing.T) {
	mfs := newMemFS()
	_ = mfs.WriteFile("/cache/good", []byte("data"), 0o644)
	p := Provider{Dir: "/cache", FS: mfs}
	if _, err := p.Read(context.Background(), "/abs"); err == nil {
		t.Fatalf("expected error")
	}
}
