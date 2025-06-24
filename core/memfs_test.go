package core

import (
	"io/fs"
	"os"
	"testing"
)

type memFS struct{ files map[string]memFile }

type memFile struct {
	data []byte
	mode fs.FileMode
}

func newMemFS() *memFS {
	return &memFS{files: make(map[string]memFile)}
}

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

func useMemFS(t *testing.T) *memFS {
	m := newMemFS()
	origRead, origWrite := readFile, writeFile
	readFile = m.ReadFile
	writeFile = m.WriteFile
	t.Cleanup(func() {
		readFile = origRead
		writeFile = origWrite
	})
	return m
}
