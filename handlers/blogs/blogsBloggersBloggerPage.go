package blogs

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
)

func BloggersBloggerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		//Rows []*GetCountOfBlogPostsByUserRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
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

	handlers.TemplateHandler(w, r, "bloggersBloggerPage.gohtml", data)
}
