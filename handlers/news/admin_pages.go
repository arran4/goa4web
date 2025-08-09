package news

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminNewsPage(w http.ResponseWriter, r *http.Request) {
	type RoleInfo struct {
		ID       int32
		Username sql.NullString
		Email    string
		Roles    []string
	}
	type Data struct {
		Error     string
		CanPost   bool
		UserRoles []RoleInfo
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	ps := cd.PageSize()
	cd.NextLink = fmt.Sprintf("/admin/news?offset=%d", offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("/admin/news?offset=%d", offset-ps)
		cd.StartLink = "/admin/news?offset=0"
	}
	cd.PageTitle = "News Admin"

	data := Data{Error: r.URL.Query().Get("error"), CanPost: cd.HasGrant("news", "post", "edit", 0) && cd.AdminMode}
	queries := cd.Queries()
	users, err := queries.AdminListAllUsers(r.Context())
	if err == nil {
		userMap := make(map[int32]*RoleInfo)
		for _, u := range users {
			userMap[u.Idusers] = &RoleInfo{ID: u.Idusers, Username: u.Username, Email: u.Email}
		}
		if rows, err := queries.GetUserRoles(r.Context()); err == nil {
			for _, row := range rows {
				u := userMap[row.UsersIdusers]
				if u == nil {
					u = &RoleInfo{ID: row.UsersIdusers, Username: row.Username, Email: row.Email}
					userMap[row.UsersIdusers] = u
				}
				u.Roles = append(u.Roles, row.Role)
			}
		}
		for _, u := range userMap {
			data.UserRoles = append(data.UserRoles, *u)
		}
		sort.Slice(data.UserRoles, func(i, j int) bool {
			return data.UserRoles[i].Username.String < data.UserRoles[j].Username.String
		})
	}

	if err := cd.ExecuteSiteTemplate(w, r, "adminNewsListPage.gohtml", data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}

func AdminNewsPostPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Post           *db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
		TopicID        int32
		Thread         *db.GetThreadLastPosterAndPermsRow
		Comments       []*db.GetCommentsByThreadIdForUserRow
		IsReplyable    bool
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		http.Redirect(w, r, "/admin/news", http.StatusTemporaryRedirect)
		return
	}
	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: cd.UserID,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		http.Redirect(w, r, "/admin/news?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	topicID, err := queries.GetForumTopicIdByThreadId(r.Context(), post.ForumthreadID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetForumTopicIdByThreadId: %v", err)
	}

	comments, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: cd.UserID,
		ThreadID: int32(post.ForumthreadID),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetCommentsByThreadIdForUser: %v", err)
	}
	threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      cd.UserID,
		ThreadID:      int32(post.ForumthreadID),
		ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetThreadLastPosterAndPerms: %v", err)
	}

	cd.PageTitle = fmt.Sprintf("News Post %d", pid)
	data := Data{
		Post:        post,
		TopicID:     topicID,
		Thread:      threadRow,
		Comments:    comments,
		IsReplyable: false,
	}
	data.CanEditComment = func(*db.GetCommentsByThreadIdForUserRow) bool { return false }
	data.EditURL = func(*db.GetCommentsByThreadIdForUserRow) string { return "" }
	data.EditSaveURL = func(*db.GetCommentsByThreadIdForUserRow) string { return "" }
	data.Editing = func(*db.GetCommentsByThreadIdForUserRow) bool { return false }
	data.AdminURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if cd.HasRole("administrator") {
			return fmt.Sprintf("/admin/comment/%d", cmt.Idcomments)
		}
		return ""
	}

	if err := cd.ExecuteSiteTemplate(w, r, "adminNewsPostPage.gohtml", data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}

func adminNewsEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		http.Redirect(w, r, "/admin/news", http.StatusTemporaryRedirect)
		return
	}
	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: cd.UserID,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		http.Redirect(w, r, "/admin/news?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	langs, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	cd.PageTitle = "Edit News"
	data := struct {
		Languages          []*db.Language
		Post               *db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
		SelectedLanguageId int
	}{
		Languages:          langs,
		Post:               post,
		SelectedLanguageId: int(post.LanguageIdlanguage.Int32),
	}
	if err := cd.ExecuteSiteTemplate(w, r, "adminNewsEditPage.gohtml", data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}

func AdminNewsDeleteConfirmPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		http.Redirect(w, r, "/admin/news", http.StatusTemporaryRedirect)
		return
	}
	cd.PageTitle = "Confirm news delete"
	data := struct {
		PostID       int
		ConfirmLabel string
		Back         string
	}{
		PostID:       pid,
		ConfirmLabel: "Confirm delete",
		Back:         fmt.Sprintf("/admin/news/article/%d", pid),
	}
	if err := cd.ExecuteSiteTemplate(w, r, "adminNewsDeleteConfirmPage.gohtml", data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
