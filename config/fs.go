package config

import (
	"io/fs"
	"os"
)

// FileSystem abstracts basic file operations for configuration helpers.
type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

// OSFS implements FileSystem using the os package.
type OSFS struct{}

func (OSFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
func (OSFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
