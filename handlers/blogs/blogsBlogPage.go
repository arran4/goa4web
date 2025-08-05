package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/internal/db"

	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func BlogPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*db.GetBlogEntryForListerByIDRow
		EditUrl     string
		IsReplyable bool
	}
	type BlogComment struct {
		*db.GetCommentsByThreadIdForUserRow
		ShowReply bool
		EditUrl   string
		Editing   bool
		Offset    int
		Idblogs   int32
	}
	type Data struct {
		*common.CoreData
		Blog               *BlogRow
		Comments           []*BlogComment
		Offset             int
		IsReplyable        bool
		Text               string
		EditUrl            string
		Languages          []*db.Language
		SelectedLanguageId int
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		CoreData:           cd,
		Offset:             offset,
		IsReplyable:        true,
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
		EditUrl:            fmt.Sprintf("/blogs/blog/%d/edit", blogId),
	}

	languageRows, err := data.CoreData.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Languages = languageRows

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	blog, err := queries.GetBlogEntryForListerByID(r.Context(), db.GetBlogEntryForListerByIDParams{
		ListerID: uid,
		ID:       int32(blogId),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err == nil {
		if blog.Username.Valid {
			cd.PageTitle = fmt.Sprintf("Blog by %s", blog.Username.String)
		} else {
			cd.PageTitle = fmt.Sprintf("Blog %d", blog.Idblogs)
		}
	}
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := templates.GetCompiledSiteTemplates(r.Context().Value(consts.KeyCoreData).(*common.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", struct{}{}); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return
		default:
			log.Printf("getBlogEntryForListerByID_comments Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	if !data.CoreData.HasGrant("blogs", "entry", "view", blog.Idblogs) {
		if err := templates.GetCompiledSiteTemplates(r.Context().Value(consts.KeyCoreData).(*common.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", struct{}{}); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}

	editUrl := ""
	if uid == blog.UsersIdusers {
		editUrl = fmt.Sprintf("/blogs/blog/%d/edit", blog.Idblogs)
	}

	data.Blog = &BlogRow{
		GetBlogEntryForListerByIDRow: blog,
		EditUrl:                      editUrl,
		IsReplyable:                  true,
	}

	if !blog.ForumthreadID.Valid {
		data.IsReplyable = false
		data.Blog.IsReplyable = false
	} else {
		threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
			ViewerID:      uid,
			ThreadID:      blog.ForumthreadID.Int32,
			ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("GetThreadLastPosterAndPerms: %v", err)
			}
			data.IsReplyable = false
			data.Blog.IsReplyable = false
		} else if threadRow.Locked.Valid && threadRow.Locked.Bool {
			data.IsReplyable = false
			data.Blog.IsReplyable = false
		}

		rows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
			ViewerID: uid,
			ThreadID: blog.ForumthreadID.Int32,
			UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("getCommentsByThreadIdForUser Error: %s", err)
				handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
				return
			}
		}

		for i, row := range rows {
			editUrl := ""
			if data.CoreData.CanEditAny() || row.IsOwner {
				editUrl = fmt.Sprintf("/blogs/blog/%d/comments?comment=%d#edit", blog.Idblogs, row.Idcomments)
			}
			data.Comments = append(data.Comments, &BlogComment{
				GetCommentsByThreadIdForUserRow: row,
				ShowReply:                       true,
				EditUrl:                         editUrl,
				Offset:                          i + offset,
				Idblogs:                         blog.Idblogs,
			})
		}
	}

	handlers.TemplateHandler(w, r, "blogPage.gohtml", data)
}
