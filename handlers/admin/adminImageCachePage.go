package admin

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload"
)

// ImageCacheEntry represents a cached image entry for the admin listing.
type ImageCacheEntry struct {
	ID   string
	Size int64
}

// AdminImageCachePage lists cached images and offers bulk maintenance actions.
func AdminImageCachePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Image Cache"

	entries, totalSize, err := listImageCacheEntries(cd.Config.ImageCacheDir)
	if err != nil {
		log.Printf("list image cache: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	_, canPrune := upload.CacheProviderFromConfig(cd.Config).(upload.CacheProvider)
	maxBytes := int64(cd.Config.ImageCacheMaxBytes)
	hasMax := maxBytes > 0
	if !hasMax {
		maxBytes = 0
	}

	type Data struct {
		Entries    []ImageCacheEntry
		TotalSize  int64
		MaxSize    int64
		HasMax     bool
		CanPrune   bool
		TaskList   string
		TaskPrune  string
		TaskDelete string
	}

	data := Data{
		Entries:    entries,
		TotalSize:  totalSize,
		MaxSize:    maxBytes,
		HasMax:     hasMax,
		CanPrune:   canPrune && hasMax,
		TaskList:   string(TaskImageCacheList),
		TaskPrune:  string(TaskImageCachePrune),
		TaskDelete: string(TaskImageCacheDelete),
	}

	AdminImageCachePageTmpl.Handle(w, r, data)
}

func listImageCacheEntries(dir string) ([]ImageCacheEntry, int64, error) {
	if dir == "" {
		return nil, 0, nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	if !info.IsDir() {
		return nil, 0, fmt.Errorf("cache dir is not a directory")
	}
	var entries []ImageCacheEntry
	var total int64
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		id := filepath.Base(path)
		if !intimages.ValidID(id) {
			return nil
		}
		entries = append(entries, ImageCacheEntry{ID: id, Size: info.Size()})
		total += info.Size()
		return nil
	}); err != nil {
		return nil, total, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries, total, nil
}

// AdminImageCachePageTmpl renders the admin image cache page.
const AdminImageCachePageTmpl tasks.Template = "admin/imageCachePage.gohtml"
