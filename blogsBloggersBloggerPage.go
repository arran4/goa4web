package main

import (
	"log"
	"net/http"
)

func blogsBloggersBloggerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		//Rows []*show_blogger_listRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	//queries := r.Context().Value(ContextValues("queries")).(*Queries)
	//
	//rows, err := queries.show_blogger_list(r.Context())
	//if err != nil {
	//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	return
	//}
	//data.Rows = rows

	CustomBlogIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsBloggersBloggerPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
