package forum

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminTopicsPage shows all forum topics for management.
func AdminTopicsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Topics     []*db.Forumtopic
		Categories []*db.Forumcategory
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Topics"
	queries := cd.Queries()

	topics, err := queries.GetAllForumTopics(r.Context())
	if err != nil {
		log.Printf("GetAllForumTopics: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	cats, err := queries.GetAllForumCategories(r.Context())
	if err != nil {
		log.Printf("GetAllForumCategories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData:   cd,
		Topics:     topics,
		Categories: cats,
	}

	handlers.TemplateHandler(w, r, "adminTopicsPage.gohtml", data)
}

func AdminTopicEditPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if err := queries.UpdateForumTopic(r.Context(), db.UpdateForumTopicParams{
		Title:                        sql.NullString{String: name, Valid: true},
		Description:                  sql.NullString{String: desc, Valid: true},
		ForumcategoryIdforumcategory: int32(cid),
		Idforumtopic:                 int32(tid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/admin/forum/topics", http.StatusTemporaryRedirect)
}

func AdminTopicCreatePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if _, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
		ForumcategoryIdforumcategory: int32(pcid),
		Title:                        sql.NullString{String: name, Valid: true},
		Description:                  sql.NullString{String: desc, Valid: true},
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/admin/forum/topics", http.StatusTemporaryRedirect)
}

func AdminTopicDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if err := queries.AdminDeleteForumTopic(r.Context(), int32(tid)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/admin/forum/topics", http.StatusTemporaryRedirect)
}
