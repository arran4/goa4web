package search

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchForumTask struct{ tasks.TaskString }

var searchForumTask = &SearchForumTask{TaskString: TaskSearchForum}
var _ tasks.Task = (*SearchForumTask)(nil)

func (SearchForumTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		*common.CoreData
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !common.CanSearch(cd, "forum") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}
	data := Data{
		CoreData: cd,
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	if comments, emptyWords, noResults, err := ForumCommentSearchNotInRestrictedTopic(w, r, queries, uid); err != nil {
		return nil
	} else {
		data.Comments = comments
		data.CommentsNoResults = emptyWords
		data.CommentsEmptyWords = noResults
	}

	return handlers.TemplateWithDataHandler("resultForumActionPage.gohtml", data)
}

func ForumCommentSearchNotInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries db.Querier, uid int32) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var commentIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.CommentsSearchFirstNotInRestrictedTopic(r.Context(), db.CommentsSearchFirstNotInRestrictedTopicParams{
				ListerID: uid,
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				UserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("commentsSearchFirst Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			commentIds = ids
		} else {
			ids, err := queries.CommentsSearchNextNotInRestrictedTopic(r.Context(), db.CommentsSearchNextNotInRestrictedTopicParams{
				ListerID: uid,
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids:    commentIds,
				UserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("commentsSearchNext Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return nil, false, false, err
		}
	}

	return comments, false, false, nil
}

func ForumCommentSearchInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries db.Querier, forumTopicId []int32, uid int32) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var commentIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.CommentsSearchFirstInRestrictedTopic(r.Context(), db.CommentsSearchFirstInRestrictedTopicParams{
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
					log.Printf("commentsSearchFirst Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			commentIds = ids
		} else {
			ids, err := queries.CommentsSearchNextInRestrictedTopic(r.Context(), db.CommentsSearchNextInRestrictedTopicParams{
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

					log.Printf("commentsSearchNext Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return nil, false, false, err
		}
	}

	return comments, false, false, nil
}
