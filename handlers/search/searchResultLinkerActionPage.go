package search

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	hlinker "github.com/arran4/goa4web/handlers/linker"
	db "github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchLinkerTask struct{ tasks.TaskString }

var searchLinkerTask = &SearchLinkerTask{TaskString: TaskSearchLinker}
var _ tasks.Task = (*SearchLinkerTask)(nil)

func (SearchLinkerTask) Action(w http.ResponseWriter, r *http.Request) {
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

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
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

	handlers.TemplateHandler(w, r, "resultLinkerActionPage.gohtml", data)
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
