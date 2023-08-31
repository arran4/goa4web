package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
)

func imagebbsAdminNewBoardPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Boards []*showAllBoardsRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	boardRows, err := queries.showAllBoards(r.Context())
	if err != nil {
		log.Printf("showAllBoards Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Boards = boardRows

	CustomImageBBSIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "imagebbsAdminNewBoardPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func imagebbsAdminNewBoardMakePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	parentBoardId, _ := strconv.Atoi(r.PostFormValue("pbid"))

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	err := queries.makeImageBoard(r.Context(), makeImageBoardParams{
		ImageboardIdimageboard: int32(parentBoardId),
		Title:                  sql.NullString{Valid: true, String: name},
		Description:            sql.NullString{Valid: true, String: desc},
	})
	if err != nil {
		log.Printf("Error: makeImageBoard: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/imagebbs/admin/boards", http.StatusTemporaryRedirect)
}
