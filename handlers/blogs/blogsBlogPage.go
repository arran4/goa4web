package blogs

import (
	"database/sql"
	"errors"
	"fmt"

	db "github.com/arran4/goa4web/internal/db"

	"log"
	"net/http"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	hcommon "github.com/arran4/goa4web/handlers/common"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func BlogPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*db.GetBlogEntryForUserByIdRow
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
		*CoreData
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

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(hcommon.KeyCoreData).(*CoreData),
		Offset:             offset,
		IsReplyable:        true,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
		EditUrl:            fmt.Sprintf("/blogs/blog/%d/edit", blogId),
	}

	languageRows, err := data.CoreData.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	blog, err := queries.GetBlogEntryForUserById(r.Context(), db.GetBlogEntryForUserByIdParams{
		ViewerIdusers: uid,
		ID:            int32(blogId),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_ = templates.GetCompiledTemplates(r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
			return
		default:
			log.Printf("getBlogEntryForUserById_comments Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	if !data.CoreData.HasGrant("blogs", "entry", "view", blog.Idblogs) {
		_ = templates.GetCompiledTemplates(r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
		return
	}

	editUrl := ""
	if uid == blog.UsersIdusers {
		editUrl = fmt.Sprintf("/blogs/blog/%d/edit", blog.Idblogs)
	}

	data.Blog = &BlogRow{
		GetBlogEntryForUserByIdRow: blog,
		EditUrl:                    editUrl,
		IsReplyable:                true,
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
	}

	hcommon.TemplateHandler(w, r, "blogPage.gohtml", data)
}
