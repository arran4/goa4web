package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
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
const AdminImageCacheDetailsPageTmpl tasks.Template = "admin/imageCacheDetailsPage.gohtml"

type AdminImageCacheDetailsPage struct{}

func (p *AdminImageCacheDetailsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	id := mux.Vars(r)["id"]

	if !intimages.ValidID(id) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid cache id"))
		return
	}

	cd.PageTitle = fmt.Sprintf("Image Cache Details: %s", id)

	fullPath, err := getCachePath(cd.Config.ImageCacheDir, id)
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

	if strings.HasSuffix(base, "_thumb") {
		parentID = strings.TrimSuffix(base, "_thumb") + ext
		thumbID = id
	} else {
		parentID = id
		thumbID = base + "_thumb" + ext
	}

	parentExists := false
	if parentID != id {
		if p, err := getCachePath(cd.Config.ImageCacheDir, parentID); err == nil {
			if _, err := os.Stat(p); err == nil {
				parentExists = true
			}
		}
	}

	thumbExists := false
	if thumbID != id {
		if p, err := getCachePath(cd.Config.ImageCacheDir, thumbID); err == nil {
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

	AdminImageCacheDetailsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminImageCacheDetailsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Image Cache Details", "", &AdminImageCachePage{}
}

func (p *AdminImageCacheDetailsPage) PageTitle() string {
	return "Image Cache Details"
}

var _ common.Page = (*AdminImageCacheDetailsPage)(nil)
var _ http.Handler = (*AdminImageCacheDetailsPage)(nil)

func getCachePath(dir, id string) (string, error) {
	if !intimages.ValidID(id) {
		return "", fmt.Errorf("invalid cache id")
	}
	sub1, sub2 := id[:2], id[2:4]
	return filepath.Join(dir, sub1, sub2, id), nil
}
