package forum

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core"
)

func ThreadPageWithBasePath(w http.ResponseWriter, r *http.Request, basePath string) {
	type Data struct {
		Category       *ForumcategoryPlus
		Topic          *ForumtopicPlus
		Thread         *db.GetThreadLastPosterAndPermsRow
		Comments       []*db.GetCommentsByThreadIdForUserRow
		IsReplyable    bool
		Text           string
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
		CanReply       bool
		BasePath       string
		Labels         []templates.TopicLabel
		BackURL        string
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cd.ForumBasePath = basePath
	common.WithOffset(offset)(cd)
	data := Data{
		IsReplyable: true,
		BasePath:    basePath,
		BackURL:     r.URL.RequestURI(),
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
	displayTitle := topicRow.Title.String
	if topicRow.Handler == "private" && cd.Queries() != nil {
		parts, err := cd.Queries().ListPrivateTopicParticipantsByTopicIDForUser(r.Context(), db.ListPrivateTopicParticipantsByTopicIDForUserParams{
			TopicID:  sql.NullInt32{Int32: topicRow.Idforumtopic, Valid: true},
			ViewerID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			log.Printf("list private participants: %v", err)
		} else {
			var names []string
			for _, p := range parts {
				if p.Idusers != cd.UserID {
					names = append(names, p.Username.String)
				}
			}
			if len(names) > 0 {
				displayTitle = strings.Join(names, ", ")
			}
		}
	}
	cd.PageTitle = fmt.Sprintf("Forum - %s", displayTitle)

	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	commentRows, err := cd.SelectedThreadComments()
	if err != nil {
		log.Printf("thread comments: %v", err)
	}

	// threadRow and topicRow are provided by the RequireThreadAndTopic
	// middleware.

	commentId, _ := strconv.Atoi(r.URL.Query().Get("comment"))
	quoteId, _ := strconv.Atoi(r.URL.Query().Get("quote"))
	data.Comments = commentRows

	data.CanEditComment = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return cmt.IsOwner
	}
	data.EditURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("%s/topic/%d/thread/%d?comment=%d#edit", data.BasePath, topicRow.Idforumtopic, threadRow.Idforumthread, cmt.Idcomments)
	}
	data.EditSaveURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("%s/topic/%d/thread/%d/comment/%d", data.BasePath, topicRow.Idforumtopic, threadRow.Idforumthread, cmt.Idcomments)
	}
	data.Editing = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return data.CanEditComment(cmt) && commentId != 0 && int32(commentId) == cmt.Idcomments
	}
	data.AdminURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if cd.IsAdmin() && cd.IsAdminMode() {
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
		DisplayTitle:                 displayTitle,
		Threads:                      topicRow.Threads,
		Comments:                     topicRow.Comments,
		Lastaddition:                 topicRow.Lastaddition,
		Lastposterusername:           topicRow.Lastposterusername,
		Edit:                         false,
		Labels:                       nil,
	}

	var labels []templates.TopicLabel
	if pub, author, err := cd.ThreadPublicLabels(threadRow.Idforumthread); err == nil {
		for _, l := range pub {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "public"})
		}
		for _, l := range author {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "author"})
		}
	} else {
		log.Printf("list public labels: %v", err)
	}
	if priv, err := cd.ThreadPrivateLabels(threadRow.Idforumthread); err == nil {
		for _, l := range priv {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "private"})
		}
	} else {
		log.Printf("list private labels: %v", err)
	}
	sort.Slice(labels, func(i, j int) bool { return labels[i].Name < labels[j].Name })
	data.Labels = labels

	replyType := r.URL.Query().Get("type")
	if quoteId != 0 {
		if c, err := cd.CommentByID(int32(quoteId)); err == nil && c != nil {
			switch replyType {
			case "full":
				data.Text = a4code.QuoteText(c.Username.String, c.Text.String, a4code.WithFullQuote())
			default:
				data.Text = a4code.QuoteText(c.Username.String, c.Text.String)
			}
		}
	}

	handlers.TemplateHandler(w, r, "threadPage.gohtml", data)
}

// ThreadPage serves the forum thread page at the default /forum prefix.
func ThreadPage(w http.ResponseWriter, r *http.Request) {
	ThreadPageWithBasePath(w, r, "/forum")
}
