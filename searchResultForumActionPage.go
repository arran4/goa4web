package goa4web

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
)

func searchResultForumActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Comments           []*GetCommentsByIdsForUserWithThreadInfoRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}
	queries := r.Context().Value(common.KeyQueries).(*Queries)
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

	if err := templates.RenderTemplate(w, "resultForumActionPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func ForumCommentSearchNotInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries *Queries, uid int32) ([]*GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := common.BreakupTextToWords(r.PostFormValue("searchwords"))
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
			ids, err := queries.CommentsSearchNextNotInRestrictedTopic(r.Context(), CommentsSearchNextNotInRestrictedTopicParams{
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

	comments, err := queries.GetCommentsByIdsForUserWithThreadInfo(r.Context(), GetCommentsByIdsForUserWithThreadInfoParams{
		UsersIdusers: uid,
		Ids:          commentIds,
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

func ForumCommentSearchInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries *Queries, forumTopicId []int32, uid int32) ([]*GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := common.BreakupTextToWords(r.PostFormValue("searchwords"))
	var commentIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.CommentsSearchFirstInRestrictedTopic(r.Context(), CommentsSearchFirstInRestrictedTopicParams{
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
			ids, err := queries.CommentsSearchNextInRestrictedTopic(r.Context(), CommentsSearchNextInRestrictedTopicParams{
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

	comments, err := queries.GetCommentsByIdsForUserWithThreadInfo(r.Context(), GetCommentsByIdsForUserWithThreadInfoParams{
		UsersIdusers: uid,
		Ids:          commentIds,
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
