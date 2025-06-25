package writings

import (
	"database/sql"
	"errors"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
)

var writingsPermissionsPageEnabled = true

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Categories                       []*db.Writingcategory
		EditingCategoryId                int32
		CategoryId                       int32
		WritingcategoryIdwritingcategory int32
		IsAdmin                          bool
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
	}

	data.IsAdmin = data.CoreData.HasRole("administrator")
	editID, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	data.EditingCategoryId = int32(editID)
	data.CategoryId = 0
	data.WritingcategoryIdwritingcategory = data.CategoryId

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	categoryRows, err := queries.GetAllWritingCategories(r.Context(), data.CategoryId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllWritingCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Categories = categoryRows

	CustomWritingsIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "page.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomWritingsIndex(data *corecommon.CoreData, r *http.Request) {
	data.CustomIndexItems = append(data.CustomIndexItems,
		corecommon.IndexItem{Name: "Atom Feed", Link: "/writings/atom"},
		corecommon.IndexItem{Name: "RSS Feed", Link: "/writings/rss"},
	)
	data.RSSFeedUrl = "/writings/rss"
	data.AtomFeedUrl = "/writings/atom"

	userHasAdmin := data.HasRole("administrator")
	if userHasAdmin && writingsPermissionsPageEnabled {
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
			Name: "User Permissions",
			Link: "/writings/user/permissions",
		})
	}
	userHasWriter := data.HasRole("writer")
	if userHasWriter || userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
			Name: "Write writings",
			Link: "/writings/add",
		})
	}

	data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
		Name: "Return to list",
		Link: fmt.Sprintf("/writings?offset=%d", 0),
	})
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset != 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
			Name: "The start",
			Link: fmt.Sprintf("/writings?offset=%d", 0),
		})
	}
	data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
		Name: "Next 10",
		Link: fmt.Sprintf("/writings?offset=%d", offset+10),
	})
	if offset > 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
			Name: "Previous 10",
			Link: fmt.Sprintf("/writings?offset=%d", offset-10),
		})
	}
}
