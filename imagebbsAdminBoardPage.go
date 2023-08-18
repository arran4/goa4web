package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func imagebbsAdminBoardModifyBoardActionPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	parentBoardId, _ := strconv.Atoi(r.PostFormValue("pbid"))
	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["board"])

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	err := queries.changeImageBoard(r.Context(), changeImageBoardParams{
		ImageboardIdimageboard: int32(parentBoardId),
		Title:                  sql.NullString{Valid: true, String: name},
		Description:            sql.NullString{Valid: true, String: desc},
		Idimageboard:           int32(bid),
	})
	if err != nil {
		log.Printf("Error: makeImageBoard: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/imagebbs/admin/boards", http.StatusTemporaryRedirect)
}
