package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func imagebbsBoardPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Boards      []*Imageboard
		IsSubBoard  bool
		BoardNumber int
		Posts       []*GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountRow
	}

	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])

	data := Data{
		CoreData:    r.Context().Value(ContextValues("coreData")).(*CoreData),
		IsSubBoard:  bid != 0,
		BoardNumber: bid,
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	subBoardRows, err := queries.GetAllBoardsByParentBoardId(r.Context(), int32(bid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllBoardsByParentBoardId Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Boards = subBoardRows

	posts, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCount(r.Context(), int32(bid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllBoardsByParentBoardId Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Posts = posts

	CustomImageBBSIndex(data.CoreData, r)

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "imagebbsBoardPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func imagebbsBoardPostImageActionPage(w http.ResponseWriter, r *http.Request) {
	thumbnailURL := r.PostFormValue("thumbnailURL")
	fullimageURL := r.PostFormValue("fullimageURL")
	text := r.PostFormValue("text")

	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])

	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := queries.CreateImagePost(r.Context(), CreateImagePostParams{
		ImageboardIdimageboard: int32(bid),
		Thumbnail:              sql.NullString{Valid: true, String: thumbnailURL},
		Fullimage:              sql.NullString{Valid: true, String: fullimageURL},
		UsersIdusers:           uid,
		Description:            sql.NullString{Valid: true, String: text},
	}); err != nil {
		log.Printf("printTopicRestrictions Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// TODO not searchable: updateSearch

	taskDoneAutoRefreshPage(w, r)
}
