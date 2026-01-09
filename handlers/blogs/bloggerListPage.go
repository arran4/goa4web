package blogs

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"

	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/handlers/share"

	"github.com/arran4/goa4web/handlers"
)

// BloggerListPage shows all bloggers with their post counts.
func BloggerListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Rows     []*db.ListBloggersForListerRow
		Search   string
		PageSize int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Bloggers"
	cd.OpenGraph = &common.OpenGraph{
		Title:       cd.PageTitle,
		Description: "List of bloggers",
		Image:       share.MakeImageURL(cd.AbsoluteURL(""), cd.PageTitle, cd.ShareSigner, time.Now().Add(24*time.Hour)),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
	}
	data := Data{
		Search:   r.URL.Query().Get("search"),
		PageSize: cd.PageSize(),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pageSize := cd.PageSize()
	rows, err := cd.Bloggers(r)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	data.Rows = rows

	if data.Search != "" {
		if len(rows) == 1 {
			http.Redirect(w, r, "/blogs/blogger/"+rows[0].Username.String, http.StatusSeeOther)
			return
		}
		if len(rows) == 0 {
			cd.PageTitle = fmt.Sprintf("No bloggers found for %q", data.Search)
		} else {
			cd.PageTitle = fmt.Sprintf("Bloggers matching %q", data.Search)
		}
	}

	base := "/blogs/bloggers"
	if data.Search != "" {
		base += "?search=" + url.QueryEscape(data.Search)
	}
	if hasMore {
		if strings.Contains(base, "?") {
			cd.NextLink = fmt.Sprintf("%s&offset=%d", base, offset+pageSize)
		} else {
			cd.NextLink = fmt.Sprintf("%s?offset=%d", base, offset+pageSize)
		}
	}
	if offset > 0 {
		if strings.Contains(base, "?") {
			cd.PrevLink = fmt.Sprintf("%s&offset=%d", base, offset-pageSize)
		} else {
			cd.PrevLink = fmt.Sprintf("%s?offset=%d", base, offset-pageSize)
		}
	}

	handlers.TemplateHandler(w, r, "bloggerListPage.gohtml", data)
}
