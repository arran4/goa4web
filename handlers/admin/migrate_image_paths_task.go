package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// MigrateImagePathsTask removes the "uploads/" prefix from image paths.
type MigrateImagePathsTask struct{ tasks.TaskString }

var migrateImagePathsTask = &MigrateImagePathsTask{TaskString: "migrate_image_paths"}

var _ tasks.Task = (*MigrateImagePathsTask)(nil)
var _ tasks.AuditableTask = (*MigrateImagePathsTask)(nil)

func (MigrateImagePathsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	images, err := cd.Queries().AdminListAllUploadedImages(r.Context())
	if err != nil {
		return fmt.Errorf("list images: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	updatedCount := 0
	for _, img := range images {
		if !img.Path.Valid {
			continue
		}
		path := img.Path.String
		if strings.HasPrefix(path, "uploads/") {
			newPath := strings.TrimPrefix(path, "uploads/")
			if err := cd.Queries().AdminUpdateUploadedImagePath(r.Context(), db.AdminUpdateUploadedImagePathParams{
				Path:            sql.NullString{String: newPath, Valid: true},
				Iduploadedimage: img.Iduploadedimage,
			}); err != nil {
				return fmt.Errorf("update image %d: %w", img.Iduploadedimage, handlers.ErrRedirectOnSamePageHandler(err))
			}
			updatedCount++
		}
	}

	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["MigratedImagesCount"] = updatedCount
	}

	return handlers.RefreshDirectHandler{TargetURL: "/admin/images/uploads"}
}

// AuditRecord summarises the migration action.
func (MigrateImagePathsTask) AuditRecord(data map[string]any) string {
	if count, ok := data["MigratedImagesCount"].(int); ok {
		return fmt.Sprintf("migrated %d image paths from uploads/ prefix", count)
	}
	return "migrated image paths"
}
