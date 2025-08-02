package imagebbs

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	type Data struct {
		*common.CoreData
		Stats []*db.AdminImageboardPostCountsRow
	}
	var stats []*db.AdminImageboardPostCountsRow
	if s, err := cd.Queries().AdminImageboardPostCounts(r.Context()); err == nil {
		stats = s
	} else if err != sql.ErrNoRows {
		log.Printf("imagebbsAdminPage stats: %v", err)
	}
	cd.PageTitle = "Image Board Admin"
	data := Data{CoreData: cd, Stats: stats}
	handlers.TemplateHandler(w, r, "imagebbsAdminPage", data)
}
