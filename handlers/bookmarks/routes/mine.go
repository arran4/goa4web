package routes

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/handlers/bookmarks/internal"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func MinePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "My Bookmarks"

	var cols []*internal.Column
	bm, err := cd.Bookmarks()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("error getBookmarksForUser: %s", err)
	} else if bm != nil {
		list := strings.TrimSpace(bm.List.String)
		if list != "" {
			cols = internal.ParseColumns(list)
		}
	}

	MinePageTmpl.Handle(w, r, struct {
		Columns []*internal.Column
	}{cols})
}

const MinePageTmpl tasks.Template = "bookmarks/minePage.gohtml"
