package forum

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/blogs"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core"
)

func ThreadPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Category            *ForumcategoryPlus
		Topic               *ForumtopicPlus
		Thread              *db.GetThreadLastPosterAndPermsRow
		Comments            []*db.GetCommentsByThreadIdForUserRow
		IsReplyable         bool
		Text                string
		CategoryBreadcrumbs []*ForumcategoryPlus
		CanEditComment      func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL             func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL         func(*db.GetCommentsByThreadIdForUserRow) string
		Editing             func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL            func(*db.GetCommentsByThreadIdForUserRow) string
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	common.WithOffset(offset)(cd)
	data := Data{
		CoreData:    cd,
		IsReplyable: true,
	}

	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		log.Printf("current thread: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		log.Printf("current topic: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum - %s", topicRow.Title.String)

	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	commentRows, err := cd.SelectedThreadComments()
	if err != nil {
		log.Printf("thread comments: %v", err)
	}

	// threadRow and topicRow are provided by the RequireThreadAndTopic
	// middleware.

	data.Topic = &ForumtopicPlus{
		Idforumtopic:                 topicRow.Idforumtopic,
		Lastposter:                   topicRow.Lastposter,
		ForumcategoryIdforumcategory: topicRow.ForumcategoryIdforumcategory,
		Title:                        topicRow.Title,
		Description:                  topicRow.Description,
		Threads:                      topicRow.Threads,
		Comments:                     topicRow.Comments,
		Lastaddition:                 topicRow.Lastaddition,
		Lastposterusername:           topicRow.Lastposterusername,
		Edit:                         false,
	}

	commentId, _ := strconv.Atoi(r.URL.Query().Get("comment"))
	data.Comments = commentRows

	data.CanEditComment = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return data.CoreData.CanEditAny() || cmt.IsOwner
	}
	data.EditURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("/forum/topic/%d/thread/%d?comment=%d#edit", topicRow.Idforumtopic, threadRow.Idforumthread, cmt.Idcomments)
	}
	data.EditSaveURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("/forum/topic/%d/thread/%d/comment/%d", topicRow.Idforumtopic, threadRow.Idforumthread, cmt.Idcomments)
	}
	data.Editing = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return data.CanEditComment(cmt) && commentId != 0 && int32(commentId) == cmt.Idcomments
	}
	data.AdminURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if cd.HasRole("administrator") {
			return fmt.Sprintf("/admin/comment/%d", cmt.Idcomments)
		}
		return ""
	}
	if commentId != 0 {
		data.IsReplyable = false
	}

	data.Thread = threadRow
	data.Topic = &ForumtopicPlus{
		Idforumtopic:                 topicRow.Idforumtopic,
		Lastposter:                   topicRow.Lastposter,
		ForumcategoryIdforumcategory: topicRow.ForumcategoryIdforumcategory,
		Title:                        topicRow.Title,
		Description:                  topicRow.Description,
		Threads:                      topicRow.Threads,
		Comments:                     topicRow.Comments,
		Lastaddition:                 topicRow.Lastaddition,
		Lastposterusername:           topicRow.Lastposterusername,
		Edit:                         false,
	}

	categoryRows, err := data.CoreData.ForumCategories()
	if err != nil {
		log.Printf("getAllForumCategories Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	categoryTree := NewCategoryTree(categoryRows, []*ForumtopicPlus{data.Topic})
	data.CategoryBreadcrumbs = categoryTree.CategoryRoots(int32(topicRow.ForumcategoryIdforumcategory))

	replyType := r.URL.Query().Get("type")
	if c, err := cd.CurrentComment(r); err == nil && c != nil {
		data.IsReplyable = false
		switch replyType {
		case "full":
			data.Text = a4code.FullQuoteOf(c.Username.String, c.Text.String)
		default:
			data.Text = a4code.QuoteOfText(c.Username.String, c.Text.String)
		}
	}

	blogs.CustomBlogIndex(data.CoreData, r)

	handlers.TemplateHandler(w, r, "threadPage.gohtml", data)
}
