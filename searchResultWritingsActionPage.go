package goa4web

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
)

func searchResultWritingsActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Comments           []*GetCommentsByIdsForUserWithThreadInfoRow
		Writings           []*GetWritingsByIdsForUserDescendingByPublishedDateRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
		NoResults          bool
		EmptyWords         bool
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	ftbn, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{Valid: true, String: WritingTopicName})
	if err != nil {
		log.Printf("findForumTopicByTitle Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if comments, emptyWords, noResults, err := ForumCommentSearchInRestrictedTopic(w, r, queries, []int32{ftbn.Idforumtopic}, uid); err != nil {
		return
	} else {
		data.Comments = comments
		data.CommentsNoResults = emptyWords
		data.CommentsEmptyWords = noResults
	}

	if writings, emptyWords, noResults, err := WritingSearch(w, r, queries, uid); err != nil {
		return
	} else {
		data.Writings = writings
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	if err := templates.RenderTemplate(w, "resultWritingsActionPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func WritingSearch(w http.ResponseWriter, r *http.Request, queries *Queries, uid int32) ([]*GetWritingsByIdsForUserDescendingByPublishedDateRow, bool, bool, error) {
	searchWords := breakupTextToWords(r.PostFormValue("searchwords"))
	var writingsIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.WritingSearchFirst(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("writingSearchFirst Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			writingsIds = ids
		} else {
			ids, err := queries.WritingSearchNext(r.Context(), WritingSearchNextParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: writingsIds,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("writingSearchNext Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			writingsIds = ids
		}
		if len(writingsIds) == 0 {
			return nil, false, true, nil
		}
	}

	writings, err := queries.GetWritingsByIdsForUserDescendingByPublishedDate(r.Context(), GetWritingsByIdsForUserDescendingByPublishedDateParams{
		Userid:     uid,
		Writingids: writingsIds,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getWritings Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return nil, false, false, err
		}
	}

	return writings, false, false, nil
}
