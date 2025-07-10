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
	"time"

	"github.com/arran4/goa4web/core/templates"
)

func AdminUsersRestrictionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		MaxUserLevel    int32
		UserTopicLevels []*db.GetAllForumTopicsWithPermissionsAndTopicRow
		Users           []*db.AllUsersRow
		Topics          []*db.Forumtopic
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

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

	if err := templates.RenderTemplate(w, "adminUsersRestrictionsPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminUsersRestrictionsUpdatePage(w http.ResponseWriter, r *http.Request) {
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
	expStr := r.PostFormValue("expiresAt")
	var expires sql.NullTime
	if expStr != "" {
		t, err := time.Parse("2006-01-02", expStr)
		if err != nil {
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		expires = sql.NullTime{Time: t, Valid: true}
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := queries.UpsertUsersForumTopicLevelPermission(r.Context(), db.UpsertUsersForumTopicLevelPermissionParams{
		Level: sql.NullInt32{
			Valid: true,
			Int32: int32(level),
		},
		Invitemax: sql.NullInt32{
			Valid: true,
			Int32: int32(inviteMax),
		},
		ExpiresAt:              expires,
		ForumtopicIdforumtopic: int32(tid),
		UsersIdusers:           int32(uid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)

}

func AdminUsersRestrictionsDeletePage(w http.ResponseWriter, r *http.Request) {
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
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := queries.DeleteUsersForumTopicLevelPermission(r.Context(), db.DeleteUsersForumTopicLevelPermissionParams{
		ForumtopicIdforumtopic: int32(tid),
		UsersIdusers:           int32(uid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)

}
