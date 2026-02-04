package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

func ShowPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Link               *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow
		Languages          []*db.Language
		SelectedLanguageId int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	data := Data{
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
	}
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])

	languageRows, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	data.Languages = languageRows

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
		ViewerID:     cd.UserID,
		ID:           int32(linkId),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	if !cd.HasGrant("linker", "link", "view", link.ID) {
		fmt.Println("TODO: FIx: Add enforced Access in router rather than task")
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	data.Link = link
	if link.Title.Valid {
		cd.PageTitle = fmt.Sprintf("Link: %s", link.Title.String)
	} else {
		cd.PageTitle = fmt.Sprintf("Link %d", link.ID)
	}

	LinkerShowPageTmpl.Handle(w, r, data)
}

const LinkerShowPageTmpl tasks.Template = "linker/showPage.gohtml"

func ShowReplyPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	text := r.PostFormValue("replytext")
	languageValue := r.PostFormValue("language")
	languageId, err := strconv.Atoi(languageValue)
	if err != nil {
		languageId = 0
	}

	vars := mux.Vars(r)
	linkId, err := strconv.Atoi(vars["link"])

	if err != nil {
		//redirectReplyError(w, r, err.Error(), text, languageId)
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	if linkId == 0 {
		log.Printf("Error: no bid")
		//redirectReplyError(w, r, "No bid", text, languageId)
		handlers.RedirectSeeOtherWithMessage(w, r, "", "No bid")
		return
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
		ViewerID:     cd.UserID,
		ID:           int32(linkId),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	if !cd.HasGrant("linker", "link", "view", link.ID) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	var pthid int32 = link.ThreadID
	pt, err := queries.SystemGetForumTopicByTitle(r.Context(), sql.NullString{
		String: LinkerTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopicForPoster(r.Context(), db.CreateForumTopicForPosterParams{
			ForumcategoryID: 0,
			ForumLang:       link.LanguageID,
			Title: sql.NullString{
				String: LinkerTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: LinkerTopicName,
				Valid:  true,
			},
			Handler:         "linker",
			Section:         "forum",
			GrantCategoryID: sql.NullInt32{},
			GranteeID:       sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			PosterID:        cd.UserID,
		})
		if err != nil {
			log.Printf("Error: createForumTopic: %s", err)
			//redirectReplyError(w, r, err.Error(), text, languageId)
			handlers.RedirectSeeOtherWithError(w, r, "", err)
			return
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("Error: findForumTopicByTitle: %s", err)
		//redirectReplyError(w, r, err.Error(), text, languageId)
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	} else {
		ptid = pt.Idforumtopic
	}
	if pthid == 0 {
		pthidi, err := queries.SystemCreateThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			//redirectReplyError(w, r, err.Error(), text, languageId)
			handlers.RedirectSeeOtherWithError(w, r, "", err)
			return
		}
		pthid = int32(pthidi)
		if err := queries.SystemAssignLinkerThreadID(r.Context(), db.SystemAssignLinkerThreadIDParams{
			ThreadID: pthid,
			ID:       int32(linkId),
		}); err != nil {
			log.Printf("Error: assignThreadIdToBlogEntry: %s", err)
			//redirectReplyError(w, r, err.Error(), text, languageId)
			handlers.RedirectSeeOtherWithError(w, r, "", err)
			return
		}
	}

	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/linker/show/%d", linkId)

	cid, err := cd.CreateLinkerCommentForCommenter(uid, pthid, int32(linkId), int32(languageId), text)
	if err != nil {
		log.Printf("Error: createComment: %s", err)
		//redirectReplyError(w, r, err.Error(), text, languageId)
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	if cid == 0 {
		err := handlers.ErrForbidden
		log.Printf("Error: createComment: %s", err)
		//redirectReplyError(w, r, err.Error(), text, languageId)
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}

	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             pthid,
		TopicID:              ptid,
		CommentID:            int32(cid),
		LabelItem:            "link",
		LabelItemID:          int32(linkId),
		CommentText:          text,
		CommentURL:           cd.AbsoluteURL(endUrl),
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
		IncludeSearch:        true,
	}); err != nil {
		log.Printf("linker reply side effects: %v", err)
	}

	http.Redirect(w, r, endUrl, http.StatusSeeOther)
}

func redirectReplyError(w http.ResponseWriter, r *http.Request, msg, text string, languageID int) {
	vals := url.Values{}
	for key, values := range r.URL.Query() {
		copied := make([]string, len(values))
		copy(copied, values)
		vals[key] = copied
	}
	if msg != "" {
		vals.Set("error", msg)
	}
	if text != "" {
		vals.Set("text", text)
	} else {
		vals.Del("text")
	}
	if languageID != 0 {
		vals.Set("language", strconv.Itoa(languageID))
	} else {
		vals.Del("language")
	}

	target := r.URL.Path
	if encoded := vals.Encode(); encoded != "" {
		target = target + "?" + encoded
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}
