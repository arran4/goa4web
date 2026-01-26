package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// ModifyPostTask updates an existing image post.
type ModifyPostTask struct{ tasks.TaskString }

var modifyPostTask = &ModifyPostTask{TaskString: TaskModifyPost}

var _ tasks.Task = (*ModifyPostTask)(nil)

func (ModifyPostTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	userID, _ := strconv.Atoi(vars["user"])
	board, _ := strconv.Atoi(r.PostFormValue("board"))
	desc := r.PostFormValue("desc")
	approved := r.PostFormValue("approved") == "1"
	if err := cd.Queries().AdminUpdateImagePost(r.Context(), db.AdminUpdateImagePostParams{
		ImageboardIdimageboard: sql.NullInt32{Int32: int32(board), Valid: board != 0},
		Description:            sql.NullString{Valid: true, String: desc},
		Approved:               approved,
		Idimagepost:            int32(pid),
	}); err != nil {
		return fmt.Errorf("update image post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("/admin/user/%d/imagebbs/post/%d", userID, pid))
}

// DeletePostTask removes an image post.
type DeletePostTask struct{ tasks.TaskString }

var deletePostTask = &DeletePostTask{TaskString: TaskDeletePost}

var _ tasks.Task = (*DeletePostTask)(nil)

func (DeletePostTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	userID, _ := strconv.Atoi(vars["user"])
	if err := cd.Queries().AdminDeleteImagePost(r.Context(), int32(pid)); err != nil {
		return fmt.Errorf("delete image post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("/admin/user/%d/imagebbs", userID))
}

// AdminPostEditPage displays details for a single image post and allows edits.
func AdminPostEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	if pid == 0 {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	queries := cd.Queries()
	post, err := queries.AdminGetImagePost(r.Context(), int32(pid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}
	boards, err := queries.AdminListBoards(r.Context(), db.AdminListBoardsParams{Limit: 200, Offset: 0})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Edit Image Post %d", pid)
	data := struct {
		Post   *db.AdminGetImagePostRow
		Boards []*db.Imageboard
	}{Post: post, Boards: boards}
	ImageBBSAdminPostEditPageTmpl.Handle(w, r, data)
}

const ImageBBSAdminPostEditPageTmpl tasks.Template = "imagebbs/adminPostEditPage.gohtml"

// AdminPostDashboardPage shows an overview for a single image post with links.
func AdminPostDashboardPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	if pid == 0 {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	queries := cd.Queries()
	post, err := queries.AdminGetImagePost(r.Context(), int32(pid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}
	cd.PageTitle = fmt.Sprintf("Image Post %d", pid)
	data := struct{ Post *db.AdminGetImagePostRow }{Post: post}
	ImageBBSAdminPostDashboardPageTmpl.Handle(w, r, data)
}

const ImageBBSAdminPostDashboardPageTmpl tasks.Template = "imagebbs/adminPostDashboardPage.gohtml"

// AdminPostCommentsPage lists comments for an image post's thread.
func AdminPostCommentsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	if pid == 0 {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	queries := cd.Queries()
	post, err := queries.AdminGetImagePost(r.Context(), int32(pid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}
	comments, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: cd.UserID,
		ThreadID: post.ForumthreadID,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Image Post %d Comments", pid)
	data := struct {
		Post     *db.AdminGetImagePostRow
		Comments []*db.GetCommentsByThreadIdForUserRow
	}{Post: post, Comments: comments}
	ImageBBSAdminPostCommentsPageTmpl.Handle(w, r, data)
}

const ImageBBSAdminPostCommentsPageTmpl tasks.Template = "imagebbs/adminPostCommentsPage.gohtml"
