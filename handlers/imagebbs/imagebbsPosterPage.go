package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func PosterPage(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = fmt.Sprintf("Images by %s", username)

	data, err := cd.ImageBBSPoster(username, int32(offset))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			log.Printf("ImageBBSPoster Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}

	ImageBBSPosterPageTmpl.Handle(w, r, data)
}

const ImageBBSPosterPageTmpl handlers.Page = "imagebbs/posterPage.gohtml"
