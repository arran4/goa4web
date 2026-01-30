package templates

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// TemplateSet identifies a group of embedded templates.
type TemplateSet string

const (
	// TemplateSetSite represents the site template set.
	TemplateSetSite TemplateSet = "site"
	// TemplateSetNotifications represents the notifications template set.
	TemplateSetNotifications TemplateSet = "notifications"
	// TemplateSetEmail represents the email template set.
	TemplateSetEmail TemplateSet = "email"
	// TemplateSetAssets represents the embedded static assets set.
	TemplateSetAssets TemplateSet = "assets"
)

// WriteToDir writes the embedded templates and assets to dir preserving their structure.
func WriteToDir(dir string) error {
	return WriteTemplateSetsToDir(dir, []TemplateSet{
		TemplateSetSite,
		TemplateSetNotifications,
		TemplateSetEmail,
		TemplateSetAssets,
	})
}

// WriteTemplateSetsToDir writes selected embedded template sets to dir.
func WriteTemplateSetsToDir(dir string, sets []TemplateSet) error {
	normalized, err := normalizeTemplateSets(sets)
	if err != nil {
		return err
	}
	for _, set := range normalized {
		if err := copyDir(string(set), dir); err != nil {
			return err
		}
	}
	return nil
}

// ArchiveTemplates writes selected embedded template sets to the writer using the provided format.
func ArchiveTemplates(w io.Writer, format string, sets []TemplateSet) error {
	normalized, err := normalizeTemplateSets(sets)
	if err != nil {
		return err
	}
	switch strings.ToLower(format) {
	case "zip":
		return writeZipArchive(w, normalized)
	case "tar":
		return writeTarArchive(w, normalized)
	default:
		return fmt.Errorf("unsupported archive format %q", format)
	}
}

func normalizeTemplateSets(sets []TemplateSet) ([]TemplateSet, error) {
	if len(sets) == 0 {
		return nil, fmt.Errorf("no template sets specified")
	}
	seen := make(map[TemplateSet]struct{}, len(sets))
	normalized := make([]TemplateSet, 0, len(sets))
	for _, set := range sets {
		if !isValidTemplateSet(set) {
			return nil, fmt.Errorf("unknown template set %q", set)
		}
		if _, ok := seen[set]; ok {
			continue
		}
		seen[set] = struct{}{}
		normalized = append(normalized, set)
	}
	return normalized, nil
}

func isValidTemplateSet(set TemplateSet) bool {
	switch set {
	case TemplateSetSite, TemplateSetNotifications, TemplateSetEmail, TemplateSetAssets:
		return true
	default:
		return false
	}
}

func writeZipArchive(w io.Writer, sets []TemplateSet) error {
	zw := zip.NewWriter(w)
	for _, set := range sets {
		if err := walkTemplateSet(set, func(fsys fs.FS, relPath string, d fs.DirEntry) error {
			if d.IsDir() {
				return nil
			}
			b, err := fs.ReadFile(fsys, relPath)
			if err != nil {
				return err
			}
			entry, err := zw.Create(path.Join(string(set), relPath))
			if err != nil {
				return err
			}
			_, err = entry.Write(b)
			return err
		}); err != nil {
			_ = zw.Close()
			return err
		}
	}
	return zw.Close()
}

func writeTarArchive(w io.Writer, sets []TemplateSet) error {
	tw := tar.NewWriter(w)
	for _, set := range sets {
		if err := walkTemplateSet(set, func(fsys fs.FS, relPath string, d fs.DirEntry) error {
			info, err := d.Info()
			if err != nil {
				return err
			}
			name := path.Join(string(set), relPath)
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = name
			if info.IsDir() && !strings.HasSuffix(hdr.Name, "/") {
				hdr.Name += "/"
			}
			if hdr.ModTime.IsZero() {
				hdr.ModTime = time.Now()
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			b, err := fs.ReadFile(fsys, relPath)
			if err != nil {
				return err
			}
			_, err = tw.Write(b)
			return err
		}); err != nil {
			_ = tw.Close()
			return err
		}
	}
	return tw.Close()
}

func walkTemplateSet(set TemplateSet, fn func(fsys fs.FS, relPath string, d fs.DirEntry) error) error {
	fsys, err := fs.Sub(embeddedFS, string(set))
	if err != nil {
		return err
	}
	return fs.WalkDir(fsys, ".", func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return fn(fsys, relPath, d)
	})
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
