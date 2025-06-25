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

func AdminTopicRestrictionLevelPage(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		*CoreData
		Restrictions []*GetForumTopicRestrictionsByForumTopicIdRow
	}

	queries := r.Context().Value(common.KeyQueries).(*Queries)

	data := &Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	restrictions, err := queries.GetForumTopicRestrictionsByForumTopicId(r.Context(), int32(topicId))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("printTopicRestrictions Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Restrictions = restrictions

	CustomForumIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "adminTopicRestrictionLevelPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminTopicRestrictionLevelChangePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])
	view, err := strconv.Atoi(r.PostFormValue("view"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	reply, err := strconv.Atoi(r.PostFormValue("reply"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	newthread, err := strconv.Atoi(r.PostFormValue("newthread"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	see, err := strconv.Atoi(r.PostFormValue("see"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	invite, err := strconv.Atoi(r.PostFormValue("invite"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	read, err := strconv.Atoi(r.PostFormValue("read"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	mod, err := strconv.Atoi(r.PostFormValue("mod"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	admin, err := strconv.Atoi(r.PostFormValue("admin"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	// TODO fix query / schema to overwrite existing value
	if err := queries.UpsertForumTopicRestrictions(r.Context(), UpsertForumTopicRestrictionsParams{
		ForumtopicIdforumtopic: int32(topicId),
		Viewlevel:              sql.NullInt32{Valid: true, Int32: int32(view)},
		Replylevel:             sql.NullInt32{Valid: true, Int32: int32(reply)},
		Newthreadlevel:         sql.NullInt32{Valid: true, Int32: int32(newthread)},
		Seelevel:               sql.NullInt32{Valid: true, Int32: int32(see)},
		Invitelevel:            sql.NullInt32{Valid: true, Int32: int32(invite)},
		Readlevel:              sql.NullInt32{Valid: true, Int32: int32(read)},
		Modlevel:               sql.NullInt32{Valid: true, Int32: int32(mod)},
		Adminlevel:             sql.NullInt32{Valid: true, Int32: int32(admin)},
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	notifyAdmins(r.Context(), getEmailProvider(), queries, r.URL.Path)

	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminTopicRestrictionLevelDeletePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	queries := r.Context().Value(common.KeyQueries).(*Queries)

	if err := queries.DeleteTopicRestrictionsByForumTopicId(r.Context(), int32(topicId)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	notifyAdmins(r.Context(), getEmailProvider(), queries, r.URL.Path)

	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminTopicRestrictionLevelCopyPage(w http.ResponseWriter, r *http.Request) {
	fromID, err := strconv.Atoi(r.PostFormValue("fromTopic"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	toID, err := strconv.Atoi(r.PostFormValue("toTopic"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(common.KeyQueries).(*Queries)

	src, err := queries.GetForumTopicRestrictionsByForumTopicId(r.Context(), int32(fromID))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if len(src) == 0 || !src[0].ForumtopicIdforumtopic.Valid {
		if err := queries.DeleteTopicRestrictionsByForumTopicId(r.Context(), int32(toID)); err != nil {
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	} else {
		row := src[0]
		if err := queries.UpsertForumTopicRestrictions(r.Context(), UpsertForumTopicRestrictionsParams{
			ForumtopicIdforumtopic: int32(toID),
			Viewlevel:              row.Viewlevel,
			Replylevel:             row.Replylevel,
			Newthreadlevel:         row.Newthreadlevel,
			Seelevel:               row.Seelevel,
			Invitelevel:            row.Invitelevel,
			Readlevel:              row.Readlevel,
			Modlevel:               row.Modlevel,
			Adminlevel:             row.Adminlevel,
		}); err != nil {
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	notifyAdmins(r.Context(), getEmailProvider(), queries, r.URL.Path)

	common.TaskDoneAutoRefreshPage(w, r)
}
