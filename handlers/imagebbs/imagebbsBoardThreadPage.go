package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"
)

// ReplyTask posts a reply within a thread.
type ReplyTask struct{ tasks.TaskString }

var replyTask = &ReplyTask{TaskString: TaskReply}

var _ tasks.Task = (*ReplyTask)(nil)

// ReplyTask alerts watchers of new posts and auto-subscribes the replier so
// they see further responses.
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)
var _ tasks.EmailTemplatesRequired = (*ReplyTask)(nil)

func (ReplyTask) IndexType() string { return searchworker.TypeComment }

func (ReplyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = ReplyTask{}

func (ReplyTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateImagebbsReply.EmailTemplates(), evt.Outcome == eventbus.TaskOutcomeSuccess
}

func (ReplyTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	s := NotificationTemplateImagebbsReply.NotificationTemplate()
	return &s
}

func (ReplyTask) RequiredTemplates() []tasks.Template {
	return append([]tasks.Template{tasks.Template(ImageBBSBoardThreadPageTmpl)},
		EmailTemplateImagebbsReply.RequiredTemplates()...)
}

func (ReplyTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	return string(TaskReply), evt.Path, nil
}

func (ReplyTask) AutoSubscribeGrants(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	parts := strings.Split(evt.Path, "/")
	if len(parts) >= 4 && parts[1] == "imagebbss" && parts[2] == "imagebbs" {
		if bid, err := strconv.Atoi(parts[3]); err == nil {
			return []notif.GrantRequirement{{Section: "imagebbs", Item: "board", ItemID: int32(bid), Action: "view"}}, nil
		}
	}
	return nil, nil
}

func BoardThreadPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	vars := mux.Vars(r)
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)
	thidStr := vars["thread"]
	if thidStr == "" {
		thidStr = vars["replyTo"]
	}
	thid, _ := strconv.Atoi(thidStr)
	cd.PageTitle = fmt.Sprintf("Thread %d/%d", bid, thid)

	if !cd.HasGrant("imagebbs", "board", "view", int32(bid)) {
		fmt.Println("TODO: FIx: Add enforced Access in router rather than task")
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	common.WithOffset(offset)(cd)

	data, err := cd.ImageBBSThread(int32(bid), int32(thid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		default:
			log.Printf("get image thread fail: %s", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	ImageBBSBoardThreadPageTmpl.Handle(w, r, data)
}

const ImageBBSBoardThreadPageTmpl tasks.Template = "imagebbs/boardThreadPage.gohtml"

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()

	var uid int32

	vars := mux.Vars(r)
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)

	uid, _ = session.Values["UID"].(int32)

	if bid == 0 {
		return fmt.Errorf("no bid %w", handlers.ErrRedirectOnSamePageHandler(errors.New("no bid")))
	}

	queries := cd.Queries()

	post, err := queries.GetImagePostByIDForLister(r.Context(), db.GetImagePostByIDForListerParams{
		ListerID:     uid,
		ID:           int32(bid),
		ListerUserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_ = cd.ExecuteSiteTemplate(w, r, "admin/noAccessPage.gohtml", struct{}{})
			return nil
		default:
			return fmt.Errorf("get image post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	var pthid int32 = post.ForumthreadID
	pt, err := queries.SystemGetForumTopicByTitle(r.Context(), sql.NullString{
		String: ImageBBSTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopicForPoster(r.Context(), db.CreateForumTopicForPosterParams{
			ForumcategoryID: 0,
			ForumLang:       sql.NullInt32{},
			Title: sql.NullString{
				String: ImageBBSTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: ImageBBSTopicDescription,
				Valid:  true,
			},
			Handler:         "imagebbs",
			Section:         "forum",
			GrantCategoryID: sql.NullInt32{},
			GranteeID:       sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			PosterID:        cd.UserID,
		})
		if err != nil {
			return fmt.Errorf("create forum topic fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		ptid = int32(ptidi)
	} else if err != nil {
		return fmt.Errorf("find forum topic fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	} else {
		ptid = pt.Idforumtopic
	}
	if pthid == 0 {
		pthidi, err := queries.SystemCreateThread(r.Context(), ptid)
		if err != nil {
			return fmt.Errorf("make thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		pthid = int32(pthidi)
		if err := queries.SystemAssignImagePostThreadID(r.Context(), db.SystemAssignImagePostThreadIDParams{
			ForumthreadID: pthid,
			Idimagepost:   int32(bid),
		}); err != nil {
			return fmt.Errorf("assign imagebbs to thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ = session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/imagebbss/imagebbs/%d/comments", bid)

	cid, err := cd.CreateImageBBSCommentForCommenter(uid, pthid, int32(bid), int32(languageId), text)
	if err != nil {
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cid == 0 {
		err := handlers.ErrForbidden
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             pthid,
		TopicID:              ptid,
		CommentID:            int32(cid),
		LabelItem:            "imagebbs",
		LabelItemID:          int32(bid),
		CommentText:          text,
		CommentURL:           cd.AbsoluteURL(endUrl),
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
		IncludeSearch:        true,
	}); err != nil {
		log.Printf("imagebbs reply side effects: %v", err)
	}

	return handlers.RedirectHandler(endUrl)
}
