package bookmarks

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func MinePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "My Bookmarks"

	var cols []*Column
	bm, err := cd.Bookmarks()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("error getBookmarksForUser: %s", err)
	} else if bm != nil {
		list := strings.TrimSpace(bm.List.String)
		if list != "" {
			cols = ParseColumns(list)
		}
	}

	MinePageTmpl.Handle(w, r, struct {
		Columns []*Column
	}{cols})
}

const MinePageTmpl handlers.Page = "bookmarks/minePage.gohtml"
