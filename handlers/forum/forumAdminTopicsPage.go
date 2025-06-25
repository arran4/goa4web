package forum

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func AdminTopicsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*GetAllForumCategoriesWithSubcategoryCountRow
		Topics     []*Forumtopic
	}
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	categoryRows, err := queries.GetAllForumCategoriesWithSubcategoryCount(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	topicRows, err := queries.GetAllForumTopics(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("forumTopics Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Topics = topicRows

	CustomForumIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "adminTopicsPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminTopicEditPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	if err := queries.UpdateForumTopic(r.Context(), UpdateForumTopicParams{
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

func TopicCreatePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	if _, err := queries.CreateForumTopic(r.Context(), CreateForumTopicParams{
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

func AdminTopicDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])
	if err := queries.DeleteForumTopic(r.Context(), int32(topicId)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/forum/admin/topics", http.StatusTemporaryRedirect)
}
