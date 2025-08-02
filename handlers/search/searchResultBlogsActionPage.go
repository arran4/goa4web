package search

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	hblogs "github.com/arran4/goa4web/handlers/blogs"
	"github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchBlogsTask struct{ tasks.TaskString }

var searchBlogsTask = &SearchBlogsTask{TaskString: TaskSearchBlogs}
var _ tasks.Task = (*SearchBlogsTask)(nil)

func (SearchBlogsTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		*common.CoreData
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		Blogs              []*db.Blog
		CommentsNoResults  bool
		CommentsEmptyWords bool
		NoResults          bool
		EmptyWords         bool
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !common.CanSearch(cd, "blogs") {
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

	ftbn, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{Valid: true, String: hblogs.BloggerTopicName})
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

	if blogs, emptyWords, noResults, err := BlogSearch(w, r, queries, uid); err != nil {
		return nil
	} else {
		data.Blogs = blogs
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	return handlers.TemplateWithDataHandler("resultBlogsActionPage.gohtml", data)
}

func BlogSearch(w http.ResponseWriter, r *http.Request, queries *db.Queries, uid int32) ([]*db.Blog, bool, bool, error) {
	viewerID := uid
	userID := uid
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var blogIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.BlogsSearchFirst(r.Context(), db.BlogsSearchFirstParams{
				ListerID: uid,
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				UserID: sql.NullInt32{Int32: uid, Valid: true},
			})
			if err != nil {
				log.Printf("blogsSearchFirst Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			blogIds = ids
		} else {
			ids, err := queries.BlogsSearchNext(r.Context(), db.BlogsSearchNextParams{
				ListerID: uid,
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids:    blogIds,
				UserID: sql.NullInt32{Int32: uid, Valid: true},
			})
			if err != nil {
				log.Printf("blogsSearchNext Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			blogIds = ids
		}
		if len(blogIds) == 0 {
			return nil, false, true, nil
		}
	}

	rows, err := queries.ListBlogEntriesByIDsForLister(r.Context(), db.ListBlogEntriesByIDsForListerParams{
		ListerID: viewerID,
		UserID:   sql.NullInt32{Int32: userID, Valid: userID != 0},
		Blogids:  blogIds,
	})
	if err != nil {
		log.Printf("getBlogEntriesByIdsDescending Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, false, false, err
	}
	blogs := make([]*db.Blog, 0, len(rows))
	for _, r := range rows {
		blogs = append(blogs, &db.Blog{
			Idblogs:            r.Idblogs,
			ForumthreadID:      r.ForumthreadID,
			UsersIdusers:       r.UsersIdusers,
			LanguageIdlanguage: r.LanguageIdlanguage,
			Blog:               r.Blog,
			Written:            r.Written,
		})
	}

	return blogs, false, false, nil
}
