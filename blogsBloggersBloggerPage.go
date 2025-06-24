package goa4web

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func blogsBloggersBloggerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		//Rows []*GetCountOfBlogPostsByUserRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	//queries := r.Context().Name(ContextValues("queries")).(*Queries)
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

	if err := templates.RenderTemplate(w, "bloggersBloggerPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
