package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
)

func forumAdminTopicsRestrictionLevelPage(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		*CoreData
		Restrictions []*getAllTopicRestrictionsRow
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	data := &Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	restrictions, err := queries.getAllTopicRestrictions(r.Context())
	if err != nil {
		log.Printf("printTopicRestrictions Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Restrictions = restrictions

	CustomForumIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "forumAdminTopicsRestrictionLevelPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func forumAdminTopicsRestrictionLevelChangePage(w http.ResponseWriter, r *http.Request) {
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
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := queries.setTopicRestrictions(r.Context(), setTopicRestrictionsParams{
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

	// TODO notify admin

	taskDoneAutoRefreshPage(w, r)
}

func forumAdminTopicsRestrictionLevelDeletePage(w http.ResponseWriter, r *http.Request) {
	ftid, err := strconv.Atoi(r.PostFormValue("ftid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := queries.deleteTopicRestrictions(r.Context(), int32(ftid)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	// TODO notify admin

	taskDoneAutoRefreshPage(w, r)
}
