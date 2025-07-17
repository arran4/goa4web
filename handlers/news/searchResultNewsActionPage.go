package news

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/internal/searchworker"
)

func SearchResultNewsActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		News               []*db.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow
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

	ftbn, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{Valid: true, String: NewsTopicName})
	if err != nil {
		log.Printf("findForumTopicByTitle Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if comments, emptyWords, noResults, err := forumCommentSearchInRestrictedTopic(w, r, queries, []int32{ftbn.Idforumtopic}, uid); err != nil {
		return
	} else {
		data.Comments = comments
		data.CommentsNoResults = emptyWords
		data.CommentsEmptyWords = noResults
	}

	if news, emptyWords, noResults, err := NewsSearch(w, r, queries, uid); err != nil {
		return
	} else {
		data.News = news
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	handlers.TemplateHandler(w, r, "resultNewsActionPage.gohtml", data)
}

func NewsSearch(w http.ResponseWriter, r *http.Request, queries *db.Queries, uid int32) ([]*db.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var newsIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.SiteNewsSearchFirst(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("newsSearchFirst Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			newsIds = ids
		} else {
			ids, err := queries.SiteNewsSearchNext(r.Context(), db.SiteNewsSearchNextParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: newsIds,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("newsSearchNext Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			newsIds = ids
		}
		if len(newsIds) == 0 {
			return nil, false, true, nil
		}
	}

	news, err := queries.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountParams{
		ViewerID: uid,
		Newsids:  newsIds,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getNews Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return nil, false, false, err
		}
	}

	return news, false, false, nil
}

func forumCommentSearchInRestrictedTopic(w http.ResponseWriter, r *http.Request, queries *db.Queries, forumTopicId []int32, uid int32) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var commentIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.CommentsSearchFirstInRestrictedTopic(r.Context(), db.CommentsSearchFirstInRestrictedTopicParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ftids: forumTopicId,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("commentsSearchFirst Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			commentIds = ids
		} else {
			ids, err := queries.CommentsSearchNextInRestrictedTopic(r.Context(), db.CommentsSearchNextInRestrictedTopicParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids:   commentIds,
				Ftids: forumTopicId,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					log.Printf("commentsSearchNext Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return nil, false, false, err
				}
			}
			commentIds = ids
		}
		if len(commentIds) == 0 {
			return nil, false, true, nil
		}
	}

	comments, err := queries.GetCommentsByIdsForUserWithThreadInfo(r.Context(), db.GetCommentsByIdsForUserWithThreadInfoParams{
		ViewerID: uid,
		Ids:      commentIds,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getCommentsByIds Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return nil, false, false, err
		}
	}

	return comments, false, false, nil
}
