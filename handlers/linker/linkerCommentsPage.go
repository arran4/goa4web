package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

func CommentsPage(w http.ResponseWriter, r *http.Request) {
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
		Link               *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow
		CanReply           bool
		Languages          []*db.Language
		Comments           []*CommentPlus
		SelectedLanguageId int
		IsReplyable        bool
		Offset             int
		Text               string
		CanEdit            bool
		UserId             int32
		Thread             *db.GetThreadLastPosterAndPermsRow
	}

	offset := 0
	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = off
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	data := Data{
		CoreData:           cd,
		CanReply:           cd.UserID != 0,
		CanEdit:            false,
		Offset:             offset,
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
	}
	vars := mux.Vars(r)
	linkId := 0
	if lid, err := strconv.Atoi(vars["link"]); err == nil {
		linkId = lid
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	queries = r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
		ViewerID:     cd.UserID,
		Idlinker:     int32(linkId),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return
		default:
			log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if !(cd.HasGrant("linker", "link", "view", link.Idlinker) ||
		cd.HasGrant("linker", "link", "comment", link.Idlinker) ||
		cd.HasGrant("linker", "link", "reply", link.Idlinker)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	data.Link = link
	data.CoreData.PageTitle = fmt.Sprintf("Link %d Comments", link.Idlinker)
	data.CanEdit = cd.HasRole("administrator") || uid == link.UsersIdusers

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: uid,
		ThreadID: link.ForumthreadID,
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

	threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      link.ForumthreadID,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Error: getThreadByIdForUserByIdWithLastPosterUserNameAndPermissions: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = off
	}

	commentIdString := r.URL.Query().Get("comment")
	var commentId int
	if cid, err := strconv.Atoi(commentIdString); err == nil {
		commentId = cid
	}
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if data.CoreData.CanEditAny() || row.IsOwner {
			editUrl = fmt.Sprintf("/linker/comments/%d?comment=%d#edit", link.Idlinker, row.Idcomments)
			editSaveUrl = fmt.Sprintf("/linker/comments/%d/comment/%d", link.Idlinker, row.Idcomments)
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
			Languages:                       languageRows,
			SelectedLanguageId:              row.LanguageIdlanguage,
			AdminUrl: func() string {
				if data.CoreData.HasRole("administrator") {
					return fmt.Sprintf("/admin/comment/%d", row.Idcomments)
				} else {
					return ""
				}
			}(),
		})
	}

	data.Thread = threadRow

	handlers.TemplateHandler(w, r, "commentsPage.gohtml", data)
}

type replyTask struct{ tasks.TaskString }

var replyTaskEvent = &replyTask{TaskString: TaskReply}
var _ tasks.Task = (*replyTask)(nil)

func (replyTask) IndexType() string { return searchworker.TypeComment }

func (replyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = replyTask{}

func (replyTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}

	vars := mux.Vars(r)
	linkId, err := strconv.Atoi(vars["link"])

	if err != nil {
		return fmt.Errorf("link id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if linkId == 0 {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("no bid"))
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
		ViewerID:     cd.UserID,
		Idlinker:     int32(linkId),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return nil
		default:
			log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return nil
		}
	}

	if !(cd.HasGrant("linker", "link", "view", link.Idlinker) ||
		cd.HasGrant("linker", "link", "comment", link.Idlinker) ||
		cd.HasGrant("linker", "link", "reply", link.Idlinker)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}

	var pthid int32 = link.ForumthreadID
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: LinkerTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: LinkerTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: LinkerTopicDescription,
				Valid:  true,
			},
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
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			return fmt.Errorf("make thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		pthid = int32(pthidi)
		if err := queries.SystemAssignLinkerThreadID(r.Context(), db.SystemAssignLinkerThreadIDParams{
			ThreadID: pthid,
			LinkerID: int32(linkId),
		}); err != nil {
			return fmt.Errorf("assign linker thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	text := r.PostFormValue("replytext")
	languageId := 0
	if lid, err := strconv.Atoi(r.PostFormValue("language")); err == nil {
		languageId = lid
	}
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/linker/comments/%d", linkId)

	cid, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
		ForumthreadID:      pthid,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: pthid, TopicID: ptid}
			evt.Data["CommentURL"] = cd.AbsoluteURL(endUrl)
		}
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: int32(cid), Text: text}
		}
	}

	return handlers.RedirectHandler(endUrl)
}
