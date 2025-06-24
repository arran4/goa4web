package goa4web

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func blogsBlogAddPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
	if !(cd.HasRole("writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	type Data struct {
		*CoreData
		Languages          []*Language
		SelectedLanguageId int
		Mode               string
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(resolveDefaultLanguageID(r.Context(), queries)),
		Mode:               "Add",
	}

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomBlogIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "blogAddPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func blogsBlogAddActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	id, err := queries.CreateBlogEntry(r.Context(), CreateBlogEntryParams{
		UsersIdusers:       uid,
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

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d", id), http.StatusTemporaryRedirect)
}
