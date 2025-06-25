package blogs

import (
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func BloggersBloggerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		//Rows []*GetCountOfBlogPostsByUserRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
	}

	//queries := r.Context().Name(common.KeyQueries).(*db.Queries)
	//
	//rows, err := queries.GetCountOfBlogPostsByUser(r.Context())
	//if err != nil {
	//switch {
	//case errors.Is(err, sql.ErrNoRows):
	//default:

	//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	return
	//}
	//data.Rows = rows

	CustomBlogIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "bloggersBloggerPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
