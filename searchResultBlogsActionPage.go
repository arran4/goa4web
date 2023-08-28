package main

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func searchResultBlogsActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Comments           []*getCommentsWithThreadInfoRow
		Blogs              []*Blog
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

	ftbnId, err := queries.findForumTopicByName(r.Context(), sql.NullString{Valid: true, String: "A BLOGGER TOPIC"})
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

	if blogs, emptyWords, noResults, err := BlogSearch(w, r, queries, uid); err != nil {
		return
	} else {
		data.Blogs = blogs
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	if err := getCompiledTemplates().ExecuteTemplate(w, "searchResultBlogsActionPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func BlogSearch(w http.ResponseWriter, r *http.Request, queries *Queries, uid int32) ([]*Blog, bool, bool, error) {
	searchWords := breakupTextToWords(r.PostFormValue("searchwords"))
	var blogIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.blogsSearchFirst(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				log.Printf("blogsSearchFirst Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			blogIds = ids
		} else {
			ids, err := queries.blogsSearchNext(r.Context(), blogsSearchNextParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: blogIds,
			})
			if err != nil {
				log.Printf("blogsSearchNext Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			blogIds = ids
		}
		if len(blogIds) == 0 {
			return nil, false, true, nil
		}
	}

	blogs, err := queries.getBlogs(r.Context(), blogIds)
	if err != nil {
		log.Printf("getBlogs Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, false, false, err
	}

	return blogs, false, false, nil
}
