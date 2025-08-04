package forum

import (
	"database/sql"
	"errors"
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
	type CommentPlus struct {
		*db.GetCommentsByThreadIdForUserRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Languages          []*db.Language
		SelectedLanguageId int32
		EditSaveUrl        string
		AdminUrl           string
	}
	type Data struct {
		*common.CoreData
		Category            *ForumcategoryPlus
		Topic               *ForumtopicPlus
		Thread              *db.GetThreadLastPosterAndPermsRow
		Comments            []*CommentPlus
		Offset              int
		IsReplyable         bool
		Text                string
		Languages           []*db.Language
		SelectedLanguageId  int
		CategoryBreadcrumbs []*ForumcategoryPlus
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	data := Data{
		CoreData:           cd,
		Offset:             offset,
		IsReplyable:        true,
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
	}

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		log.Printf("current thread: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		log.Printf("current topic: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum - %s", topicRow.Title.String)

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: uid,
		ThreadID: threadRow.Idforumthread,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getBlogEntryForListerByID_comments Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
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

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if data.CoreData.CanEditAny() || row.IsOwner {
			editUrl = fmt.Sprintf("/forum/topic/%d/thread/%d?comment=%d#edit", topicRow.Idforumtopic, threadRow.Idforumthread, row.Idcomments)
			editSaveUrl = fmt.Sprintf("/forum/topic/%d/thread/%d/comment/%d", topicRow.Idforumtopic, threadRow.Idforumthread, row.Idcomments)
			if commentId != 0 && int32(commentId) == row.Idcomments {
				data.IsReplyable = false
			}
		}

		data.Comments = append(data.Comments, &CommentPlus{
			GetCommentsByThreadIdForUserRow: row,
			ShowReply:                       true,
			EditUrl:                         editUrl,
			EditSaveUrl:                     editSaveUrl,
			Editing:                         commentId != 0 && int32(commentId) == row.Idcomments,
			Offset:                          i + offset,
			Languages:                       nil,
			SelectedLanguageId:              0,
			AdminUrl: func() string {
				if cd.HasRole("administrator") {
					return fmt.Sprintf("/admin/comment/%d", row.Idcomments)
				} else {
					return ""
				}
			}(),
		})
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
	if commentIdString != "" {
		comment, err := queries.GetCommentByIdForUser(r.Context(), db.GetCommentByIdForUserParams{
			ViewerID: uid,
			ID:       int32(commentId),
			UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			log.Printf("getCommentByIdForUser Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		switch replyType {
		case "full":
			data.Text = a4code.FullQuoteOf(comment.Username.String, comment.Text.String)
		default:
			data.Text = a4code.QuoteOfText(comment.Username.String, comment.Text.String)
		}
	}

	blogs.CustomBlogIndex(data.CoreData, r)

	handlers.TemplateHandler(w, r, "threadPage.gohtml", data)
}
