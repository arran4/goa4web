package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func forumAdminUserLevelPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		MaxUserLevel    int32
		UserTopicLevels []*getUsersAllTopicLevelsRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	uid, _ := strconv.Atoi(vars["user"])

	rows, err := queries.getUsersAllTopicLevels(r.Context(), int32(uid))
	if err != nil {
		log.Printf("getAllUsersTopicLevels Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.UserTopicLevels = rows

	CustomForumIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "forumAdminUserLevelPage.tmpl", data); err != nil {
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

	if !(existsTopicRestrictions(cont, topicid) && (level == auth_administrator || topicAllowAdmin(cont, topicid))) {
		printf("Unable to view document or file.<br>\n")
	} else if level != auth_administrator && getUsersTopicLevel(cont, topicid, uid) >= getUsersTopicLevel(cont, topicid, cont.user.UID) {
		printf("Unable to modify someone with a higher level then your own.<br>\n")
	} else if level != auth_administrator && getUsersTopicLevel(cont, topicid, uid) >= getUsersTopicLevelInviteMax(cont, topicid, cont.user.UID) {
		printf("Unable to modify someone with a higher level then your have been allowed to modify.<br>\n")
	}

	if err := queries.setUsersTopicLevel(r.Context(), setUsersTopicLevelParams{
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

func forumAdminUserLevelDeletePage(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	uid, _ := strconv.Atoi(vars["user"])

	if !(existsTopicRestrictions(cont, topicid) && (level == auth_administrator || topicAllowAdmin(cont, topicid))) {
		printf("Unable to view document or file.<br>\n")
	} else if level != auth_administrator && getUsersTopicLevel(cont, topicid, uid) >= getUsersTopicLevel(cont, topicid, cont.user.UID) {
		printf("Unable to modify someone with a higher level then your own.<br>\n")
	} else if level != auth_administrator && getUsersTopicLevel(cont, topicid, uid) >= getUsersTopicLevelInviteMax(cont, topicid, cont.user.UID) {
		printf("Unable to modify someone with a higher level then your have been allowed to modify.<br>\n")
	}

	if err := queries.deleteUsersTopicLevel(r.Context(), deleteUsersTopicLevelParams{
		ForumtopicIdforumtopic: int32(tid),
		UsersIdusers:           int32(uid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	// TODO notify admin

	taskDoneAutoRefreshPage(w, r)

}
