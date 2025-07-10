package core

import (
	"io/fs"
	"os"
)

// FileSystem abstracts basic file operations so tests can use an in-memory implementation.
type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

// DirFS extends FileSystem with directory operations used by startup checks.
type DirFS interface {
	FileSystem
	MkdirAll(path string, perm fs.FileMode) error
	Stat(name string) (fs.FileInfo, error)
	Remove(name string) error
}

// OSFS implements FileSystem using the os package.
type OSFS struct{}

func (OSFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
func (OSFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// OSDirFS implements DirFS using the os package.
type OSDirFS struct{ OSFS }

func (OSDirFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
func (OSDirFS) Stat(name string) (fs.FileInfo, error)        { return os.Stat(name) }
func (OSDirFS) Remove(name string) error                     { return os.Remove(name) }
