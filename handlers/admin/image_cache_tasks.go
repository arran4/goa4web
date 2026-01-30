package admin

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload"
)

// ImageCacheListTask refreshes the cache listing view.
type ImageCacheListTask struct{ tasks.TaskString }

var imageCacheListTask = &ImageCacheListTask{TaskString: TaskImageCacheList}

var _ tasks.Task = (*ImageCacheListTask)(nil)

func (ImageCacheListTask) Action(http.ResponseWriter, *http.Request) any {
	return handlers.RefreshDirectHandler{TargetURL: "/admin/images/cache"}
}

// ImageCachePruneTask prunes the cache to the configured maximum size.
type ImageCachePruneTask struct{ tasks.TaskString }

var imageCachePruneTask = &ImageCachePruneTask{TaskString: TaskImageCachePrune}

var _ tasks.Task = (*ImageCachePruneTask)(nil)
var _ tasks.AuditableTask = (*ImageCachePruneTask)(nil)

func (ImageCachePruneTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cp := upload.CacheProviderFromConfig(cd.Config); cp != nil {
		if ccp, ok := cp.(upload.CacheProvider); ok {
			if err := ccp.Cleanup(r.Context(), int64(cd.Config.ImageCacheMaxBytes)); err != nil {
				return fmt.Errorf("prune cache: %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["ImageCachePruned"] = true
		evt.Data["ImageCacheMaxBytes"] = cd.Config.ImageCacheMaxBytes
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/images/cache"}
}

// AuditRecord summarises the prune action.
func (ImageCachePruneTask) AuditRecord(data map[string]any) string {
	if max, ok := data["ImageCacheMaxBytes"].(int); ok && max > 0 {
		return fmt.Sprintf("pruned image cache to max %d bytes", max)
	}
	return "pruned image cache"
}

// ImageCacheDeleteTask deletes selected cached images.
type ImageCacheDeleteTask struct{ tasks.TaskString }

var imageCacheDeleteTask = &ImageCacheDeleteTask{TaskString: TaskImageCacheDelete}

var _ tasks.Task = (*ImageCacheDeleteTask)(nil)
var _ tasks.AuditableTask = (*ImageCacheDeleteTask)(nil)

func (ImageCacheDeleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, id := range r.Form["id"] {
		full, err := cachePath(cd.Config.ImageCacheDir, id)
		if err != nil {
			return fmt.Errorf("invalid cache id: %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("delete cache file: %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["ImageCacheDeletedIDs"] = appendString(evt.Data["ImageCacheDeletedIDs"], id)
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/images/cache"}
}

// AuditRecord summarises deleted cache entries.
func (ImageCacheDeleteTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["ImageCacheDeletedIDs"].([]string); ok && len(ids) > 0 {
		return fmt.Sprintf("deleted %d image cache files (%s)", len(ids), strings.Join(ids, ","))
	}
	return "deleted image cache files"
}

func cachePath(dir, id string) (string, error) {
	if !intimages.ValidID(id) {
		return "", fmt.Errorf("invalid cache id")
	}
	sub1, sub2 := id[:2], id[2:4]
	return filepath.Join(dir, sub1, sub2, id), nil
}

func appendString(existing any, id string) []string {
	if ids, ok := existing.([]string); ok {
		return append(ids, id)
	}
	return []string{id}
}
