package admin

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestImageCacheMaintenanceResolvesVersionedFiles(t *testing.T) {
	// Listing and removal exercise the real local cache directory layout.
	dir := t.TempDir()
	id := "abcd1234.png"
	otherID := "abcdefgh.png"
	digest := strings.Repeat("a", 64)
	shard := filepath.Join(dir, "ab", "cd")
	legacy := filepath.Join(shard, id)
	published := filepath.Join(shard, digest, id)
	other := filepath.Join(shard, digest, otherID)
	for _, name := range []string{legacy, published, other} {
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Happy Path", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.ImageCacheDir = dir
		q := testhelpers.NewQuerierStub()
		q.GetImageCacheEntryReturns = &db.ImageCacheEntry{
			ID: id, Checksum: sql.NullString{String: "sha256:" + digest, Valid: true},
		}
		app := common.NewCoreData(t.Context(), q, cfg)

		name, _, err := getCachePath(t.Context(), app, id)
		if err != nil || name != published {
			t.Fatalf("published path = %q, err = %v", name, err)
		}
		entries, total, err := listImageCacheEntries(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 2 || total != 12 || entries[0] != (ImageCacheEntry{ID: id, Size: 8}) {
			t.Fatalf("entries = %#v, total = %d", entries, total)
		}
	})

	t.Run("Delete selected entry", func(t *testing.T) {
		if err := removeImageCacheFiles(dir, id); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{legacy, published} {
			if _, err := os.Stat(name); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("deleted path %q stat error = %v", name, err)
			}
		}
		if _, err := os.Stat(other); err != nil {
			t.Fatalf("unselected image was removed: %v", err)
		}
	})
}
