package main

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func searchResultWritingsActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Comments           []*getCommentsWithThreadInfoRow
		Writings           []*fetchWritingByIdsRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
		NoResults          bool
		EmptyWords         bool
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	ftbnId, err := queries.findForumTopicByName(r.Context(), sql.NullString{Valid: true, String: "A WRITINGS TOPIC"})
	if err != nil {
		log.Printf("findForumTopicByName Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if comments, emptyWords, noResults, err := ForumCommentSearchInRestrictedTopic(w, r, queries, []int32{ftbnId}, uid); err != nil {
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

	if err := getCompiledTemplates().ExecuteTemplate(w, "searchResultWritingsActionPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func WritingSearch(w http.ResponseWriter, r *http.Request, queries *Queries, uid int32) ([]*fetchWritingByIdsRow, bool, bool, error) {
	searchWords := breakupTextToWords(r.PostFormValue("searchwords"))
	var writingsIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.writingSearchFirst(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				log.Printf("writingSearchFirst Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			writingsIds = ids
		} else {
			ids, err := queries.writingSearchNext(r.Context(), writingSearchNextParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: writingsIds,
			})
			if err != nil {
				log.Printf("writingSearchNext Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			writingsIds = ids
		}
		if len(writingsIds) == 0 {
			return nil, false, true, nil
		}
	}

	writings, err := queries.fetchWritingByIds(r.Context(), fetchWritingByIdsParams{
		Userid:     uid,
		Writingids: writingsIds,
	})
	if err != nil {
		log.Printf("getWritings Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, false, false, err
	}

	return writings, false, false, nil
}
