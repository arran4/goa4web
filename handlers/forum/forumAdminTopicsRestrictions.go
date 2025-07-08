package forum

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/runtimeconfig"
)

func AdminTopicsRestrictionLevelPage(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		*CoreData
		Restrictions []*db.GetAllForumTopicRestrictionWithForumTopicTitleRow
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	data := &Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	restrictions, err := queries.GetAllForumTopicRestrictionWithForumTopicTitle(r.Context())
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

	if err := templates.RenderTemplate(w, "adminTopicsRestrictionLevelPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminTopicsRestrictionLevelChangePage(w http.ResponseWriter, r *http.Request) {
	ftid, err := strconv.Atoi(r.PostFormValue("ftid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
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
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := queries.UpsertForumTopicRestriction(r.Context(), db.UpsertForumTopicRestrictionParams{
		ForumtopicIdforumtopic: int32(ftid),
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

	notifyAdmins(r.Context(), email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig), queries, r.URL.Path)

	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminTopicsRestrictionLevelDeletePage(w http.ResponseWriter, r *http.Request) {
	ftid, err := strconv.Atoi(r.PostFormValue("ftid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := queries.DeleteTopicRestrictionByForumTopicId(r.Context(), int32(ftid)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	notifyAdmins(r.Context(), email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig), queries, r.URL.Path)

	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminTopicsRestrictionLevelCopyPage(w http.ResponseWriter, r *http.Request) {
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

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	src, err := queries.GetForumTopicRestrictionByForumTopicId(r.Context(), int32(fromID))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if len(src) == 0 || !src[0].ForumtopicIdforumtopic.Valid {
		if err := queries.DeleteTopicRestrictionByForumTopicId(r.Context(), int32(toID)); err != nil {
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	} else {
		row := src[0]
		if err := queries.UpsertForumTopicRestriction(r.Context(), db.UpsertForumTopicRestrictionParams{
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

	notifyAdmins(r.Context(), email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig), queries, r.URL.Path)

	common.TaskDoneAutoRefreshPage(w, r)
}
