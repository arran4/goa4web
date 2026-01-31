package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

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

func AdminImageCacheDetailsPage(w http.ResponseWriter, r *http.Request) {
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

	link, err := cd.Queries().AdminGetExternalLinkByCacheID(r.Context(), db.AdminGetExternalLinkByCacheIDParams{
		CardImageCache: sql.NullString{String: id, Valid: true},
		FaviconCache:   sql.NullString{String: id, Valid: true},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	type Data struct {
		ID          string
		FileExists  bool
		FileSize    int64
		FullPath    string
		Link        *db.ExternalLink
		TaskRefresh string
	}

	data := Data{
		ID:          id,
		FileExists:  fileExists,
		FileSize:    fileSize,
		FullPath:    fullPath,
		Link:        link,
		TaskRefresh: string(TaskImageCacheRefresh),
	}

	AdminImageCacheDetailsPageTmpl.Handle(w, r, data)
}

func getCachePath(dir, id string) (string, error) {
	if !intimages.ValidID(id) {
		return "", fmt.Errorf("invalid cache id")
	}
	sub1, sub2 := id[:2], id[2:4]
	return filepath.Join(dir, sub1, sub2, id), nil
}
