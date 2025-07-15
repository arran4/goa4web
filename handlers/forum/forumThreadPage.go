package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/a4code"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	blogs "github.com/arran4/goa4web/handlers/blogs"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
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
	}
	type Data struct {
		*CoreData
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

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(common.KeyCoreData).(*CoreData),
		Offset:             offset,
		IsReplyable:        true,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
	}

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	threadRow := r.Context().Value(common.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(common.KeyTopic).(*db.GetForumTopicByIdForUserRow)

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
			log.Printf("getBlogEntryForUserById_comments Error: %s", err)
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

	categoryRows, err := queries.GetAllForumCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
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

	if err := templates.RenderTemplate(w, "threadPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
