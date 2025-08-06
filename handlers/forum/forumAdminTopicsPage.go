package forum

import (
	"database/sql"
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Topics"

	handlers.TemplateHandler(w, r, "adminTopicsPage.gohtml", struct{}{})
}

func AdminTopicEditPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	languageID, _ := strconv.Atoi(r.PostFormValue("language"))
	if err := cd.Queries().AdminUpdateForumTopic(r.Context(), db.AdminUpdateForumTopicParams{
		Title:                        sql.NullString{String: name, Valid: true},
		Description:                  sql.NullString{String: desc, Valid: true},
		ForumcategoryIdforumcategory: int32(cid),
		LanguageIdlanguage:           int32(languageID),
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	languageID, _ := strconv.Atoi(r.PostFormValue("language"))
	if _, err := cd.Queries().SystemCreateForumTopic(r.Context(), db.SystemCreateForumTopicParams{
		ForumcategoryIdforumcategory: int32(pcid),
		LanguageIdlanguage:           int32(languageID),
		Title:                        sql.NullString{String: name, Valid: true},
		Description:                  sql.NullString{String: desc, Valid: true},
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/admin/forum/topics", http.StatusTemporaryRedirect)
}

func AdminTopicDeletePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if err := cd.Queries().AdminDeleteForumTopic(r.Context(), int32(tid)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/admin/forum/topics", http.StatusTemporaryRedirect)
}
