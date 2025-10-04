package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"

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
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	if !cd.HasGrant("linker", "link", "view", link.ID) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	data.Link = link
	if link.Title.Valid {
		cd.PageTitle = fmt.Sprintf("Link: %s", link.Title.String)
	} else {
		cd.PageTitle = fmt.Sprintf("Link %d", link.ID)
	}

	handlers.TemplateHandler(w, r, "showPage.gohtml", data)
}

func ShowReplyPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	linkId, err := strconv.Atoi(vars["link"])

	if err != nil {
		handlers.RedirectToGet(w, r, "?error="+err.Error())
		return
	}
	if linkId == 0 {
		log.Printf("Error: no bid")
		handlers.RedirectToGet(w, r, "?error="+"No bid")
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
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
			handlers.RedirectToGet(w, r, "?error="+err.Error())
			return
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("Error: findForumTopicByTitle: %s", err)
		handlers.RedirectToGet(w, r, "?error="+err.Error())
		return
	} else {
		ptid = pt.Idforumtopic
	}
	if pthid == 0 {
		pthidi, err := queries.SystemCreateThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			handlers.RedirectToGet(w, r, "?error="+err.Error())
			return
		}
		pthid = int32(pthidi)
		if err := queries.SystemAssignLinkerThreadID(r.Context(), db.SystemAssignLinkerThreadIDParams{
			ThreadID: pthid,
			ID:       int32(linkId),
		}); err != nil {
			log.Printf("Error: assignThreadIdToBlogEntry: %s", err)
			handlers.RedirectToGet(w, r, "?error="+err.Error())
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/linker/show/%d", linkId)

	cid, err := cd.CreateLinkerCommentForCommenter(uid, pthid, int32(linkId), int32(languageId), text)
	if err != nil {
		log.Printf("Error: createComment: %s", err)
		handlers.RedirectToGet(w, r, "?error="+err.Error())
		return
	}
	if cid == 0 {
		err := handlers.ErrForbidden
		log.Printf("Error: createComment: %s", err)
		handlers.RedirectToGet(w, r, "?error="+err.Error())
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: int32(cid), ThreadID: pthid, TopicID: ptid}
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

	handlers.RedirectToGet(w, r, endUrl)
}
