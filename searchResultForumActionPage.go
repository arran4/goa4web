package main

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func searchResultForumActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Comments           []*getCommentsWithThreadInfoRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	if comments, emptyWords, noResults, err := ForumCommentSearchNotInRestrictedTopic(w, r, queries, uid); err != nil {
		return
	} else {
		data.Comments = comments
		data.CommentsNoResults = emptyWords
		data.CommentsEmptyWords = noResults
	}

	if err := getCompiledTemplates().ExecuteTemplate(w, "searchResultForumActionPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func ForumCommentSearchNotInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries *Queries, uid int32) ([]*getCommentsWithThreadInfoRow, bool, bool, error) {
	searchWords := breakupTextToWords(r.PostFormValue("searchwords"))
	var commentIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.commentsSearchFirstNotInRestrictedTopic(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				log.Printf("commentsSearchFirst Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			commentIds = ids
		} else {
			ids, err := queries.commentsSearchNextNotInRestrictedTopic(r.Context(), commentsSearchNextNotInRestrictedTopicParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: commentIds,
			})
			if err != nil {
				log.Printf("commentsSearchNext Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			commentIds = ids
		}
		if len(commentIds) == 0 {
			return nil, false, true, nil
		}
	}

	comments, err := queries.getCommentsWithThreadInfo(r.Context(), getCommentsWithThreadInfoParams{
		UsersIdusers: uid,
		Ids:          commentIds,
	})
	if err != nil {
		log.Printf("getComments Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, false, false, err
	}

	return comments, false, false, nil
}

func ForumCommentSearchInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries *Queries, forumTopicId []int32, uid int32) ([]*getCommentsWithThreadInfoRow, bool, bool, error) {
	searchWords := breakupTextToWords(r.PostFormValue("searchwords"))
	var commentIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.commentsSearchFirstInRestrictedTopic(r.Context(), commentsSearchFirstInRestrictedTopicParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ftids: forumTopicId,
			})
			if err != nil {
				log.Printf("commentsSearchFirst Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			commentIds = ids
		} else {
			ids, err := queries.commentsSearchNextInRestrictedTopic(r.Context(), commentsSearchNextInRestrictedTopicParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids:   commentIds,
				Ftids: forumTopicId,
			})
			if err != nil {
				log.Printf("commentsSearchNext Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			commentIds = ids
		}
		if len(commentIds) == 0 {
			return nil, false, true, nil
		}
	}

	comments, err := queries.getCommentsWithThreadInfo(r.Context(), getCommentsWithThreadInfoParams{
		UsersIdusers: uid,
		Ids:          commentIds,
	})
	if err != nil {
		log.Printf("getComments Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, false, false, err
	}

	return comments, false, false, nil
}
