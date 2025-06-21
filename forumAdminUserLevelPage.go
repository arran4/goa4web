package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func forumAdminUserLevelPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		MaxUserLevel    int32
		UserTopicLevels []*GetAllForumTopicsForUserWithPermissionsRestrictionsAndTopicRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	uid, _ := strconv.Atoi(vars["user"])

	rows, err := queries.GetAllForumTopicsForUserWithPermissionsRestrictionsAndTopic(r.Context(), int32(uid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllUsersTopicLevels Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.UserTopicLevels = rows

	CustomForumIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "forumAdminUserLevelPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func forumAdminUserLevelUpdatePage(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	inviteMax, err := strconv.Atoi(r.PostFormValue("inviteMax"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	level, err := strconv.Atoi(r.PostFormValue("level"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	uid, _ := strconv.Atoi(vars["user"])

	if err := queries.UpsertUsersForumTopicLevelPermission(r.Context(), UpsertUsersForumTopicLevelPermissionParams{
		Level: sql.NullInt32{
			Valid: true,
			Int32: int32(level),
		},
		Invitemax: sql.NullInt32{
			Valid: true,
			Int32: int32(inviteMax),
		},
		ForumtopicIdforumtopic: int32(tid),
		UsersIdusers:           int32(uid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	notifyAdmins(r.Context(), getEmailProvider(), queries, r.URL.Path)

	taskDoneAutoRefreshPage(w, r)

}

func forumAdminUserLevelDeletePage(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	uid, _ := strconv.Atoi(vars["user"])

	if err := queries.DeleteUsersForumTopicLevelPermission(r.Context(), DeleteUsersForumTopicLevelPermissionParams{
		ForumtopicIdforumtopic: int32(tid),
		UsersIdusers:           int32(uid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	notifyAdmins(r.Context(), getEmailProvider(), queries, r.URL.Path)

	taskDoneAutoRefreshPage(w, r)

}
