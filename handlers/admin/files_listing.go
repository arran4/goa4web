package admin

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/arran4/goa4web/handlers/imagebbs"
	"github.com/arran4/goa4web/internal/db"
)

type invalidPathError struct{}

func (invalidPathError) Error() string { return "invalid path" }

type notFoundError struct{}

func (notFoundError) Error() string { return "not found" }

// ImageFileEntry describes a stored image file and any related metadata.
type ImageFileEntry struct {
	Name     string
	Path     string
	Size     int64
	IsDir    bool
	Username string
	Board    string
	Posted   time.Time
	URL      string
	ModTime  time.Time
}

// ImageFilesListing describes a listing of image upload paths.
type ImageFilesListing struct {
	Path    string
	Parent  string
	Entries []ImageFileEntry
}

// BuildImageFilesListing returns a directory listing of stored image files.
func BuildImageFilesListing(ctx context.Context, queries db.Querier, uploadDir string, reqPath string, signKey string, sign func(id string, ttl time.Duration) string, ttl time.Duration) (ImageFilesListing, error) {
	base := filepath.Join(uploadDir, imagebbs.ImagebbsUploadPrefix)
	cleaned := filepath.Clean("/" + reqPath)
	abs := filepath.Join(base, cleaned)
	if rel, err := filepath.Rel(base, abs); err != nil || rel == ".." || strings.HasPrefix(rel, "..") {
		return ImageFilesListing{}, invalidPathError{}
	}

	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return ImageFilesListing{}, notFoundError{}
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return ImageFilesListing{}, fmt.Errorf("readdir: %w", err)
	}

	listing := ImageFilesListing{Path: cleaned}
	if cleaned != "/" {
		listing.Parent = filepath.Dir(cleaned)
	}

	for _, entry := range entries {
		fi, err := entry.Info()
		var size int64
		var modTime time.Time
		if err == nil && fi != nil {
			size = fi.Size()
			modTime = fi.ModTime()
		}
		ent := ImageFileEntry{
			Name:    entry.Name(),
			Path:    filepath.Join(cleaned, entry.Name()),
			Size:    size,
			IsDir:   entry.IsDir(),
			ModTime: modTime,
		}
		if !entry.IsDir() {
			dbPath := path.Join("/imagebbs/images", ent.Path)
			row, err := queries.GetImagePostInfoByPath(ctx, db.GetImagePostInfoByPathParams{
				Fullimage: sql.NullString{Valid: true, String: dbPath},
				Thumbnail: sql.NullString{Valid: true, String: dbPath},
			})
			if err == nil && row != nil {
				ent.Username = row.Username.String
				ent.Board = row.Title.String
				if row.Posted.Valid {
					ent.Posted = row.Posted.Time
				}
			}
			if signKey != "" && sign != nil {
				id := filepath.Base(ent.Path)
				ent.URL = sign(id, ttl)
			}
		}
		listing.Entries = append(listing.Entries, ent)
	}
	sort.Slice(listing.Entries, func(i, j int) bool { return listing.Entries[i].Name < listing.Entries[j].Name })
	return listing, nil
}
