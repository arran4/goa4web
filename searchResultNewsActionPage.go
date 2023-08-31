package main

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func searchResultNewsActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Comments           []*getCommentsWithThreadInfoRow
		News               []*getNewsPostsRow
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

	ftbnId, err := queries.findForumTopicByName(r.Context(), sql.NullString{Valid: true, String: "A NEWS TOPIC"})
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

	if news, emptyWords, noResults, err := NewsSearch(w, r, queries, uid); err != nil {
		return
	} else {
		data.News = news
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	if err := getCompiledTemplates().ExecuteTemplate(w, "searchResultNewsActionPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func NewsSearch(w http.ResponseWriter, r *http.Request, queries *Queries, uid int32) ([]*getNewsPostsRow, bool, bool, error) {
	searchWords := breakupTextToWords(r.PostFormValue("searchwords"))
	var newsIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.siteNewsSearchFirst(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				log.Printf("newsSearchFirst Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			newsIds = ids
		} else {
			ids, err := queries.siteNewsSearchNext(r.Context(), siteNewsSearchNextParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: newsIds,
			})
			if err != nil {
				log.Printf("newsSearchNext Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			newsIds = ids
		}
		if len(newsIds) == 0 {
			return nil, false, true, nil
		}
	}

	news, err := queries.getNewsPosts(r.Context(), newsIds)
	if err != nil {
		log.Printf("getNews Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, false, false, err
	}

	return news, false, false, nil
}
