package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func forumAdminUsersRestrictionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		MaxUserLevel    int32
		UserTopicLevels []*GetAllForumTopicsWithPermissionsAndTopicRow
		Users           []*User
		Topics          []*Forumtopic
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.GetAllForumTopicsWithPermissionsAndTopic(r.Context())
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

	userRows, err := queries.AllUsers(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("allUsers Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Users = userRows

	topicRows, err := queries.GetAllForumTopics(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("allTopics Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Topics = topicRows

	CustomForumIndex(data.CoreData, r)

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "forumAdminUsersRestrictionsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func forumAdminUsersRestrictionsUpdatePage(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
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

	// TODO notify admin

	taskDoneAutoRefreshPage(w, r)

}

func forumAdminUsersRestrictionsDeletePage(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := queries.DeleteUsersForumTopicLevelPermission(r.Context(), DeleteUsersForumTopicLevelPermissionParams{
		ForumtopicIdforumtopic: int32(tid),
		UsersIdusers:           int32(uid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	// TODO notify admin

	taskDoneAutoRefreshPage(w, r)

}
