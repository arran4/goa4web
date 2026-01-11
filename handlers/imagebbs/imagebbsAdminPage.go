package imagebbs

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	type Data struct {
		Stats []*db.AdminImageboardPostCountsRow
	}
	var stats []*db.AdminImageboardPostCountsRow
	if s, err := cd.Queries().AdminImageboardPostCounts(r.Context()); err == nil {
		stats = s
	} else if err != sql.ErrNoRows {
		log.Printf("imagebbsAdminPage stats: %v", err)
	}
	cd.PageTitle = "Image Board Admin"
	data := Data{Stats: stats}
	ImageBBSAdminPageTmpl.Handle(w, r, data)
}

const ImageBBSAdminPageTmpl handlers.Page = "imagebbs/adminPage.gohtml"
