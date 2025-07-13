package forum

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func AdminTopicsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*db.GetAllForumCategoriesWithSubcategoryCountRow
		Topics     []*db.Forumtopic
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

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
	if r.Method == http.MethodGet {
		AdminTopicEditFormPage(w, r)
		return
	}

	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	if err := queries.UpdateForumTopic(r.Context(), db.UpdateForumTopicParams{
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

	// Update permissions if provided
	view, _ := strconv.Atoi(r.PostFormValue("view"))
	reply, _ := strconv.Atoi(r.PostFormValue("reply"))
	newthread, _ := strconv.Atoi(r.PostFormValue("newthread"))
	see, _ := strconv.Atoi(r.PostFormValue("see"))
	invite, _ := strconv.Atoi(r.PostFormValue("invite"))
	read, _ := strconv.Atoi(r.PostFormValue("read"))
	mod, _ := strconv.Atoi(r.PostFormValue("mod"))
	admin, _ := strconv.Atoi(r.PostFormValue("admin"))
	_ = queries.UpsertForumTopicRestrictions(r.Context(), db.UpsertForumTopicRestrictionsParams{
		ForumtopicIdforumtopic: int32(topicId),
		ViewRoleID:             sql.NullInt32{Valid: true, Int32: int32(view)},
		ReplyRoleID:            sql.NullInt32{Valid: true, Int32: int32(reply)},
		NewthreadRoleID:        sql.NullInt32{Valid: true, Int32: int32(newthread)},
		SeeRoleID:              sql.NullInt32{Valid: true, Int32: int32(see)},
		InviteRoleID:           sql.NullInt32{Valid: true, Int32: int32(invite)},
		ReadRoleID:             sql.NullInt32{Valid: true, Int32: int32(read)},
		ModRoleID:              sql.NullInt32{Valid: true, Int32: int32(mod)},
		AdminRoleID:            sql.NullInt32{Valid: true, Int32: int32(admin)},
	})

	http.Redirect(w, r, "/forum/admin/topics", http.StatusTemporaryRedirect)

}

func TopicCreatePage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		TopicCreateFormPage(w, r)
		return
	}
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	tid, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
		ForumcategoryIdforumcategory: int32(pcid),
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	view, _ := strconv.Atoi(r.PostFormValue("view"))
	reply, _ := strconv.Atoi(r.PostFormValue("reply"))
	newthread, _ := strconv.Atoi(r.PostFormValue("newthread"))
	see, _ := strconv.Atoi(r.PostFormValue("see"))
	invite, _ := strconv.Atoi(r.PostFormValue("invite"))
	read, _ := strconv.Atoi(r.PostFormValue("read"))
	mod, _ := strconv.Atoi(r.PostFormValue("mod"))
	admin, _ := strconv.Atoi(r.PostFormValue("admin"))
	_ = queries.UpsertForumTopicRestrictions(r.Context(), db.UpsertForumTopicRestrictionsParams{
		ForumtopicIdforumtopic: int32(tid),
		ViewRoleID:             sql.NullInt32{Valid: true, Int32: int32(view)},
		ReplyRoleID:            sql.NullInt32{Valid: true, Int32: int32(reply)},
		NewthreadRoleID:        sql.NullInt32{Valid: true, Int32: int32(newthread)},
		SeeRoleID:              sql.NullInt32{Valid: true, Int32: int32(see)},
		InviteRoleID:           sql.NullInt32{Valid: true, Int32: int32(invite)},
		ReadRoleID:             sql.NullInt32{Valid: true, Int32: int32(read)},
		ModRoleID:              sql.NullInt32{Valid: true, Int32: int32(mod)},
		AdminRoleID:            sql.NullInt32{Valid: true, Int32: int32(admin)},
	})

	http.Redirect(w, r, "/forum/admin/topics", http.StatusTemporaryRedirect)

}

func AdminTopicEditFormPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Topic       *db.Forumtopic
		Restriction *db.GetForumTopicRestrictionsByForumTopicIdRow
		Categories  []*db.GetAllForumCategoriesWithSubcategoryCountRow
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	topic, err := queries.GetForumTopicById(r.Context(), int32(topicId))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	rrows, _ := queries.GetForumTopicRestrictionsByForumTopicId(r.Context(), int32(topicId))
	var restrict *db.GetForumTopicRestrictionsByForumTopicIdRow
	if len(rrows) > 0 {
		restrict = rrows[0]
	}
	cats, _ := queries.GetAllForumCategoriesWithSubcategoryCount(r.Context())
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*CoreData), Topic: topic, Restriction: restrict, Categories: cats}
	CustomForumIndex(data.CoreData, r)
	if err := templates.RenderTemplate(w, "adminTopicEditPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func TopicCreateFormPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*db.GetAllForumCategoriesWithSubcategoryCountRow
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cats, _ := queries.GetAllForumCategoriesWithSubcategoryCount(r.Context())
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*CoreData), Categories: cats}
	CustomForumIndex(data.CoreData, r)
	if err := templates.RenderTemplate(w, "adminTopicCreatePage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminTopicDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])
	if err := queries.DeleteForumTopic(r.Context(), int32(topicId)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/forum/admin/topics", http.StatusTemporaryRedirect)
}
