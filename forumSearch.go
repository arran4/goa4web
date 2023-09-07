package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
)

func InsertWordsToForumSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *Queries, cid int64) bool {
	for _, wid := range wordIds {
		if err := queries.AddToForumCommentSearch(r.Context(), AddToForumCommentSearchParams{
			CommentsIdcomments:             int32(cid),
			SearchwordlistIdsearchwordlist: int32(wid),
		}); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("Error: addToForumCommentSearch: %s", err)
				http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
				return true
			}
		}
	}
	return false
}
