package news

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/workers/searchworker"
)

func SearchResultNewsActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		News               []*db.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
		NoResults          bool
		EmptyWords         bool
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !common.CanSearch(cd, "news") {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	data := Data{}
	cd.PageTitle = "News Search Results"
	queries := cd.Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	ftbn, err := queries.SystemGetForumTopicByTitle(r.Context(), sql.NullString{Valid: true, String: NewsTopicName})
	if err != nil {
		log.Printf("findForumTopicByTitle Error: %s", err)
		handlers.RenderErrorPage(w, r, err)
		return
	}

	if comments, emptyWords, noResults, err := forumCommentSearchInRestrictedTopic(w, r, queries, []int32{ftbn.Idforumtopic}, uid); err != nil {
		return
	} else {
		data.Comments = comments
		data.CommentsNoResults = emptyWords
		data.CommentsEmptyWords = noResults
	}

	if news, emptyWords, noResults, err := cd.SearchNews(r, uid); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	} else {
		data.News = news
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	SearchResultNewsActionPageTmpl.Handle(w, r, data)
}

const SearchResultNewsActionPageTmpl handlers.Page = "search/resultNewsActionPage.gohtml"

func forumCommentSearchInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries db.Querier, forumTopicId []int32, uid int32) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var commentIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.ListCommentIDsBySearchWordFirstForListerInRestrictedTopic(r.Context(), db.ListCommentIDsBySearchWordFirstForListerInRestrictedTopicParams{
				ListerID: uid,
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ftids:  forumTopicId,
				UserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("ListCommentIDsBySearchWordFirstForListerInRestrictedTopic Error: %s", err)
					handlers.RenderErrorPage(w, r, err)
					return nil, false, false, err
				}
			}
			commentIds = ids
		} else {
			ids, err := queries.ListCommentIDsBySearchWordNextForListerInRestrictedTopic(r.Context(), db.ListCommentIDsBySearchWordNextForListerInRestrictedTopicParams{
				ListerID: uid,
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids:    commentIds,
				Ftids:  forumTopicId,
				UserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("ListCommentIDsBySearchWordNextForListerInRestrictedTopic Error: %s", err)
					handlers.RenderErrorPage(w, r, err)
					return nil, false, false, err
				}
			}
			commentIds = ids
		}
		if len(commentIds) == 0 {
			return nil, false, true, nil
		}
	}

	comments, err := queries.GetCommentsByIdsForUserWithThreadInfo(r.Context(), db.GetCommentsByIdsForUserWithThreadInfoParams{
		ViewerID: uid,
		Ids:      commentIds,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getCommentsByIds Error: %s", err)
			handlers.RenderErrorPage(w, r, err)
			return nil, false, false, err
		}
	}

	return comments, false, false, nil
}
