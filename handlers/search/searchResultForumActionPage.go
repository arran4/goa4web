package search

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	common "github.com/arran4/goa4web/handlers/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	searchworker "github.com/arran4/goa4web/internal/searchworker"

	"github.com/arran4/goa4web/core"
)

func SearchResultForumActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*hcommon.CoreData
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
	}

	data := Data{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
	}
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	if comments, emptyWords, noResults, err := ForumCommentSearchNotInRestrictedTopic(w, r, queries, uid); err != nil {
		return
	} else {
		data.Comments = comments
		data.CommentsNoResults = emptyWords
		data.CommentsEmptyWords = noResults
	}

	common.TemplateHandler(w, r, "resultForumActionPage.gohtml", data)
}

func ForumCommentSearchNotInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries *db.Queries, uid int32) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := searchworker.BreakupTextToWords(r.PostFormValue("searchwords"))
	var commentIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.CommentsSearchFirstNotInRestrictedTopic(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
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
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: commentIds,
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

func ForumCommentSearchInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries *db.Queries, forumTopicId []int32, uid int32) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := searchworker.BreakupTextToWords(r.PostFormValue("searchwords"))
	var commentIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.CommentsSearchFirstInRestrictedTopic(r.Context(), db.CommentsSearchFirstInRestrictedTopicParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ftids: forumTopicId,
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
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids:   commentIds,
				Ftids: forumTopicId,
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
