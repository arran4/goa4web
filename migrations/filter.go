package migrations

import (
	"io/fs"
	"path"
	"strings"
)

// FilterFS returns an fs.FS that only exposes migration files matching the specified driver.
func FilterFS(base fs.FS, driver string) fs.FS {
	if driver == "sqlite3" {
		driver = "sqlite"
	}
	return &driverFilteredFS{
		base:   base,
		driver: driver,
	}
}

// ForDriver returns the embedded migrations FS filtered for the given database driver.
func ForDriver(driver string) fs.FS {
	return FilterFS(FS, driver)
}

type driverFilteredFS struct {
	base   fs.FS
	driver string
}

func (d *driverFilteredFS) isApplicable(name string) bool {
	if !strings.HasSuffix(name, ".sql") {
		return true
	}
	base := strings.TrimSuffix(path.Base(name), ".sql")
	if strings.HasSuffix(base, "_mysql") {
		return d.driver == "mysql"
	}
	if strings.HasSuffix(base, "_sqlite") {
		return d.driver == "sqlite"
	}
	return true
}

func (d *driverFilteredFS) Open(name string) (fs.File, error) {
	if !d.isApplicable(name) {
		return nil, fs.ErrNotExist
	}
	f, err := d.base.Open(name)
	if err != nil {
		return nil, err
	}
	return &filteredFile{File: f, driver: d.driver}, nil
}

func (d *driverFilteredFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(d.base, name)
	if err != nil {
		return nil, err
	}
	var filtered []fs.DirEntry
	for _, entry := range entries {
		if entry.IsDir() || d.isApplicable(entry.Name()) {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

func (d *driverFilteredFS) ReadFile(name string) ([]byte, error) {
	if !d.isApplicable(name) {
		return nil, fs.ErrNotExist
	}
	return fs.ReadFile(d.base, name)
}

func (d *driverFilteredFS) Stat(name string) (fs.FileInfo, error) {
	if !d.isApplicable(name) {
		return nil, fs.ErrNotExist
	}
	return fs.Stat(d.base, name)
}

type filteredFile struct {
	fs.File
	driver string
}

func (f *filteredFile) ReadDir(n int) ([]fs.DirEntry, error) {
	rd, ok := f.File.(fs.ReadDirFile)
	if !ok {
		return nil, fs.ErrInvalid
	}
	entries, err := rd.ReadDir(n)
	if err != nil {
		return nil, err
	}
	var filtered []fs.DirEntry
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			filtered = append(filtered, entry)
			continue
		}
		base := strings.TrimSuffix(path.Base(name), ".sql")
		if strings.HasSuffix(base, "_mysql") {
			if f.driver == "mysql" {
				filtered = append(filtered, entry)
			}
		} else if strings.HasSuffix(base, "_sqlite") {
			if f.driver == "sqlite" {
				filtered = append(filtered, entry)
			}
		} else {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}
