package main

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func searchResultLinkerActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Comments           []*getCommentsWithThreadInfoRow
		Links              []*showLinksRow
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

	ftbnId, err := queries.findForumTopicByName(r.Context(), sql.NullString{Valid: true, String: "A LINKER TOPIC"})
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

	if Linkers, emptyWords, noResults, err := LinkerSearch(w, r, queries, uid); err != nil {
		return
	} else {
		data.Links = Linkers
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	if err := getCompiledTemplates().ExecuteTemplate(w, "searchResultLinkerActionPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func LinkerSearch(w http.ResponseWriter, r *http.Request, queries *Queries, uid int32) ([]*showLinksRow, bool, bool, error) {
	searchWords := breakupTextToWords(r.PostFormValue("searchwords"))
	var LinkerIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.linkerSearchFirst(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				log.Printf("LinkersSearchFirst Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			LinkerIds = ids
		} else {
			ids, err := queries.linkerSearchNext(r.Context(), linkerSearchNextParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: LinkerIds,
			})
			if err != nil {
				log.Printf("LinkersSearchNext Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			LinkerIds = ids
		}
		if len(LinkerIds) == 0 {
			return nil, false, true, nil
		}
	}

	Linkers, err := queries.showLinks(r.Context(), LinkerIds)
	if err != nil {
		log.Printf("getLinkers Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, false, false, err
	}

	return Linkers, false, false, nil
}
