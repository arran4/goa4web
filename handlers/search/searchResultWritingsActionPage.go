package search

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	"log"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	hwritings "github.com/arran4/goa4web/handlers/writings"
	db "github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/internal/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchWritingsTask struct{ tasks.TaskString }

var searchWritingsTask = &SearchWritingsTask{TaskString: TaskSearchWritings}

func (SearchWritingsTask) Action(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		Writings           []*db.GetWritingsByIdsForUserDescendingByPublishedDateRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
		NoResults          bool
		EmptyWords         bool
	}

	data := Data{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData),
	}
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	ftbn, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{Valid: true, String: hwritings.WritingTopicName})
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

	if writings, emptyWords, noResults, err := WritingSearch(w, r, queries, uid); err != nil {
		return
	} else {
		data.Writings = writings
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	handlers.TemplateHandler(w, r, "resultWritingsActionPage.gohtml", data)
}

func WritingSearch(w http.ResponseWriter, r *http.Request, queries *db.Queries, uid int32) ([]*db.GetWritingsByIdsForUserDescendingByPublishedDateRow, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var writingsIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.WritingSearchFirst(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("writingSearchFirst Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			writingsIds = ids
		} else {
			ids, err := queries.WritingSearchNext(r.Context(), db.WritingSearchNextParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: writingsIds,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("writingSearchNext Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			writingsIds = ids
		}
		if len(writingsIds) == 0 {
			return nil, false, true, nil
		}
	}

	writings, err := queries.GetWritingsByIdsForUserDescendingByPublishedDate(r.Context(), db.GetWritingsByIdsForUserDescendingByPublishedDateParams{
		ViewerID:      uid,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		WritingIds:    writingsIds,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getWritings Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return nil, false, false, err
		}
	}

	return writings, false, false, nil
}
