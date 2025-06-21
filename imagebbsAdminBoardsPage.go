package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
)

func imagebbsAdminBoardsPage(w http.ResponseWriter, r *http.Request) {
	type BoardRow struct {
		*Imageboard
		Threads  int32
		ModLevel int32
		Visible  bool
		Nsfw     bool
	}
	type Data struct {
		*CoreData
		Boards []*BoardRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	boardRows, err := queries.GetAllImageBoards(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllImageBoards Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	for _, b := range boardRows {
		threads, err := queries.CountThreadsByBoard(r.Context(), b.Idimageboard)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("countThreads error: %s", err)
			threads = 0
		}
		data.Boards = append(data.Boards, &BoardRow{
			Imageboard: b,
			Threads:    threads,
			ModLevel:   0,
			Visible:    true,
			Nsfw:       false,
		})
	}

	CustomImageBBSIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "imagebbsAdminBoardsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
