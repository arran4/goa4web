package search

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	hlinker "github.com/arran4/goa4web/handlers/linker"
	"github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchLinkerTask struct{ tasks.TaskString }

var searchLinkerTask = &SearchLinkerTask{TaskString: TaskSearchLinker}
var _ tasks.Task = (*SearchLinkerTask)(nil)

func (SearchLinkerTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		*common.CoreData
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		Links              []*db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
		NoResults          bool
		EmptyWords         bool
		WritingCategoryID  int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !common.CanSearch(cd, "linker") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}
	data := Data{
		CoreData: cd,
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	ftbn, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{Valid: true, String: hlinker.LinkerTopicName})
	if err != nil {
		log.Printf("findForumTopicByTitle Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil
	}

	if comments, emptyWords, noResults, err := ForumCommentSearchInRestrictedTopic(w, r, queries, []int32{ftbn.Idforumtopic}, uid); err != nil {
		return nil
	} else {
		data.Comments = comments
		data.CommentsNoResults = emptyWords
		data.CommentsEmptyWords = noResults
	}

	if Linkers, emptyWords, noResults, err := LinkerSearch(w, r, queries, uid); err != nil {
		return nil
	} else {
		data.Links = Linkers
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	return handlers.TemplateWithDataHandler("resultLinkerActionPage.gohtml", data)
}

func LinkerSearch(w http.ResponseWriter, r *http.Request, queries *db.Queries, uid int32) ([]*db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingRow, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var LinkerIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.LinkerSearchFirst(r.Context(), db.LinkerSearchFirstParams{
				ListerID: uid,
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				UserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
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
				ListerID: uid,
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids:    LinkerIds,
				UserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
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
