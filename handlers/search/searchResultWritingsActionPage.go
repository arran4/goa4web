package search

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	hwritings "github.com/arran4/goa4web/handlers/writings"
	"github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchWritingsTask struct{ tasks.TaskString }

var searchWritingsTask = &SearchWritingsTask{TaskString: TaskSearchWritings}
var _ tasks.Task = (*SearchWritingsTask)(nil)

func (SearchWritingsTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		*common.CoreData
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		Writings           []*db.ListWritingsByIDsForListerRow
		CommentsNoResults  bool
		CommentsEmptyWords bool
		NoResults          bool
		EmptyWords         bool
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !common.CanSearch(cd, "writing") {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
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

	ftbn, err := queries.SystemGetForumTopicByTitle(r.Context(), sql.NullString{Valid: true, String: hwritings.WritingTopicName})
	if err != nil {
		log.Printf("findForumTopicByTitle Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return nil
	}

	if comments, emptyWords, noResults, err := ForumCommentSearchInRestrictedTopic(w, r, queries, []int32{ftbn.Idforumtopic}, uid); err != nil {
		return nil
	} else {
		data.Comments = comments
		data.CommentsNoResults = emptyWords
		data.CommentsEmptyWords = noResults
	}

	if writings, emptyWords, noResults, err := WritingSearch(w, r, queries, cd, uid); err != nil {
		return nil
	} else {
		data.Writings = writings
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	return handlers.TemplateWithDataHandler("resultWritingsActionPage.gohtml", data)
}

func WritingSearch(w http.ResponseWriter, r *http.Request, queries db.Querier, cd *common.CoreData, uid int32) ([]*db.ListWritingsByIDsForListerRow, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var writingsIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.ListWritingSearchFirstForLister(r.Context(), db.ListWritingSearchFirstForListerParams{
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
					log.Printf("writingSearchFirst Error: %s", err)
					handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
					return nil, false, false, err
				}
			}
			writingsIds = ids
		} else {
			ids, err := queries.ListWritingSearchNextForLister(r.Context(), db.ListWritingSearchNextForListerParams{
				ListerID: uid,
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids:    writingsIds,
				UserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("writingSearchNext Error: %s", err)
					handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
					return nil, false, false, err
				}
			}
			writingsIds = ids
		}
		if len(writingsIds) == 0 {
			return nil, false, true, nil
		}
	}

	limit := int32(cd.Config.PageSizeDefault)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	writings, err := queries.ListWritingsByIDsForLister(r.Context(), db.ListWritingsByIDsForListerParams{
		ListerID:      uid,
		ListerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		WritingIds:    writingsIds,
		Limit:         limit,
		Offset:        int32(offset),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getWritings Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return nil, false, false, err
		}
	}

	return writings, false, false, nil
}
