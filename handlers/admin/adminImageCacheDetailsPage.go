package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// AdminImageCacheDetailsPageTmpl renders the admin image cache details page.
const AdminImageCacheDetailsPageTmpl tasks.Template = "domains/admin/imageCacheDetailsPage.gohtml"

func AdminImageCacheDetailsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	id := mux.Vars(r)["id"]

	if !intimages.ValidID(id) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid cache id"))
		return
	}

	cd.PageTitle = fmt.Sprintf("Image Cache Details: %s", id)

	fullPath, entry, err := getCachePath(r.Context(), cd, id)
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	fileInfo, err := os.Stat(fullPath)
	var fileSize int64
	var fileExists bool
	if err == nil {
		fileSize = fileInfo.Size()
		fileExists = true
	} else if !os.IsNotExist(err) {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	// Determine Parent and Thumb IDs
	var parentID, thumbID string
	ext := filepath.Ext(id)
	base := strings.TrimSuffix(id, ext)

	if parent := cacheParentImageID(id); parent != "" {
		parentID = parent
		thumbID = id
	} else {
		parentID = id
		thumbID = base + "_thumb" + ext
		if entry != nil && entry.ThumbnailID.Valid {
			thumbID = entry.ThumbnailID.String
		}
	}

	parentExists := false
	if parentID != id {
		if p, _, err := getCachePath(r.Context(), cd, parentID); err == nil {
			if _, err := os.Stat(p); err == nil {
				parentExists = true
			}
		}
	}

	thumbExists := false
	if thumbID != id {
		if p, _, err := getCachePath(r.Context(), cd, thumbID); err == nil {
			if _, err := os.Stat(p); err == nil {
				thumbExists = true
			}
		}
	}

	link, err := cd.Queries().AdminGetExternalLinkByCacheID(r.Context(), db.AdminGetExternalLinkByCacheIDParams{
		CardImageCache: sql.NullString{String: id, Valid: true},
		FaviconCache:   sql.NullString{String: id, Valid: true},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	// If not found and we are a thumb (or just checking parent), try parent ID
	if (link == nil || errors.Is(err, sql.ErrNoRows)) && parentID != id {
		link, err = cd.Queries().AdminGetExternalLinkByCacheID(r.Context(), db.AdminGetExternalLinkByCacheIDParams{
			CardImageCache: sql.NullString{String: parentID, Valid: true},
			FaviconCache:   sql.NullString{String: parentID, Valid: true},
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			handlers.RenderErrorPage(w, r, err)
			return
		}
	}
	// If still error (NoRows), reset link to nil
	if errors.Is(err, sql.ErrNoRows) {
		link = nil
	}

	var ownerName string
	if link != nil && link.UpdatedBy.Valid {
		user, err := cd.Queries().SystemGetUserByID(r.Context(), link.UpdatedBy.Int32)
		if err == nil && user != nil && user.Username.Valid {
			ownerName = user.Username.String
		}
	}

	type Data struct {
		ID           string
		FileExists   bool
		FileSize     int64
		FullPath     string
		Link         *db.ExternalLink
		TaskRefresh  string
		ParentID     string
		ParentExists bool
		ThumbID      string
		ThumbExists  bool
		OwnerName    string
	}

	data := Data{
		ID:           id,
		FileExists:   fileExists,
		FileSize:     fileSize,
		FullPath:     fullPath,
		Link:         link,
		TaskRefresh:  string(TaskImageCacheRefresh),
		ParentID:     parentID,
		ParentExists: parentExists,
		ThumbID:      thumbID,
		ThumbExists:  thumbExists,
		OwnerName:    ownerName,
	}

	_ = AdminImageCacheDetailsPageTmpl.Handle(w, r, data)
}

func getCachePath(ctx context.Context, cd *common.CoreData, id string) (string, *db.ImageCacheEntry, error) {
	entry, err := cd.ImageCacheEntry(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", nil, fmt.Errorf("load cache entry: %w", err)
	}
	key, err := common.ImageCacheObjectKey(id, entry)
	if err != nil {
		return "", nil, err
	}
	if parentID := cacheParentImageID(id); parentID != "" {
		parent, err := cd.ImageCacheEntry(ctx, parentID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", nil, fmt.Errorf("load parent cache entry: %w", err)
		}
		if parent != nil && (!parent.ThumbnailID.Valid || parent.ThumbnailID.String != id) {
			parentKey, err := common.ImageCacheObjectKey(parentID, parent)
			if err != nil {
				return "", nil, err
			}
			key = path.Join(path.Dir(parentKey), id)
		}
	}
	return filepath.Join(cd.Config.ImageCacheDir, filepath.FromSlash(key)), entry, nil
}

func cacheParentImageID(id string) string {
	ext := filepath.Ext(id)
	base := strings.TrimSuffix(id, ext)
	if parent, ok := strings.CutSuffix(base, "_thumb"); ok {
		return parent + ext
	}
	if index := strings.LastIndex(base, "_thumb_"); index >= 0 {
		return base[:index] + ext
	}
	return ""
}
