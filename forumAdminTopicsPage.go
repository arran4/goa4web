package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func forumAdminTopicsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*showAllCategoriesRow
		Topics     []*Forumtopic
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	categoryRows, err := queries.showAllCategories(r.Context())
	if err != nil {
		log.Printf("forumCategories Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	data.Categories = categoryRows

	topicRows, err := queries.getAllTopics(r.Context())
	if err != nil {
		log.Printf("forumTopics Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	data.Topics = topicRows

	CustomForumIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "forumAdminTopicsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func forumAdminTopicEditPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	if err := queries.changeTopic(r.Context(), changeTopicParams{
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
		Idforumtopic:                 int32(topicId),
		ForumcategoryIdforumcategory: int32(cid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/forum/admin/topics", http.StatusTemporaryRedirect)

}

func forumTopicCreatePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if _, err := queries.makeTopic(r.Context(), makeTopicParams{
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
		ForumcategoryIdforumcategory: int32(pcid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/forum/admin/topics", http.StatusTemporaryRedirect)

}
