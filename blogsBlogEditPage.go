package goa4web

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func blogsBlogEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
	if !(cd.HasRole("writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	type Data struct {
		*CoreData
		Languages          []*Language
		Blog               *GetBlogEntryForUserByIdRow
		SelectedLanguageId int
		Mode               string
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(resolveDefaultLanguageID(r.Context(), queries)),
		Mode:               "Edit",
	}

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	row, err := queries.GetBlogEntryForUserById(r.Context(), int32(blogId))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Blog = row

	CustomBlogIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "blogEditPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func blogsBlogEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	err = queries.UpdateBlogEntry(r.Context(), UpdateBlogEntryParams{
		Idblogs:            int32(blogId),
		LanguageIdlanguage: int32(languageId),
		Blog: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d", blogId), http.StatusTemporaryRedirect)
}
