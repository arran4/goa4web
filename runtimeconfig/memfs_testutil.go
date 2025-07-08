package runtimeconfig

import (
	"io/fs"
	"os"
	"testing"
)

type memFile struct {
	data []byte
	mode fs.FileMode
}

type memFS struct{ files map[string]memFile }

func newMemFS() *memFS { return &memFS{files: map[string]memFile{}} }

func (m *memFS) ReadFile(name string) ([]byte, error) {
	f, ok := m.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return append([]byte(nil), f.data...), nil
}

func (m *memFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	m.files[name] = memFile{data: append([]byte(nil), data...), mode: perm}
	return nil
}

// useMemFS returns an in-memory FileSystem for tests.
func useMemFS(t *testing.T) FileSystem {
	t.Helper()
	return newMemFS()
}
