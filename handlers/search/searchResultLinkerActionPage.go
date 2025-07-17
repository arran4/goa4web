package search

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	common "github.com/arran4/goa4web/handlers/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	hlinker "github.com/arran4/goa4web/handlers/linker"
	db "github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/internal/searchworker"

	"github.com/arran4/goa4web/core"
)

func SearchResultLinkerActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*hcommon.CoreData
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		Links              []*db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
		NoResults          bool
		EmptyWords         bool
		WritingCategoryID  int32
	}

	data := Data{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
	}
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	ftbn, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{Valid: true, String: hlinker.LinkerTopicName})
	if err != nil {
		log.Printf("findForumTopicByTitle Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if comments, emptyWords, noResults, err := ForumCommentSearchInRestrictedTopic(w, r, queries, []int32{ftbn.Idforumtopic}, uid); err != nil {
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

	common.TemplateHandler(w, r, "resultLinkerActionPage.gohtml", data)
}

func LinkerSearch(w http.ResponseWriter, r *http.Request, queries *db.Queries, uid int32) ([]*db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingRow, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var LinkerIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.LinkerSearchFirst(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("LinkersSearchFirst Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			LinkerIds = ids
		} else {
			ids, err := queries.LinkerSearchNext(r.Context(), db.LinkerSearchNextParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: LinkerIds,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("LinkersSearchNext Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			LinkerIds = ids
		}
		if len(LinkerIds) == 0 {
			return nil, false, true, nil
		}
	}

	Linkers, err := queries.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescending(r.Context(), LinkerIds)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getLinkers Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return nil, false, false, err
		}
	}

	return Linkers, false, false, nil
}
