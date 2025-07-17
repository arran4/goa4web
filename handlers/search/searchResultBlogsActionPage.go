package search

import (
	"database/sql"
	"log"
	"net/http"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	hblogs "github.com/arran4/goa4web/handlers/blogs"
	db "github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/internal/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchBlogsTask struct{ tasks.TaskString }

var searchBlogsTask = &SearchBlogsTask{TaskString: TaskSearchBlogs}

func (SearchBlogsTask) Action(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Comments           []*db.GetCommentsByIdsForUserWithThreadInfoRow
		Blogs              []*db.Blog
		CommentsNoResults  bool
		CommentsEmptyWords bool
		NoResults          bool
		EmptyWords         bool
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	ftbn, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{Valid: true, String: hblogs.BloggerTopicName})
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

	if blogs, emptyWords, noResults, err := BlogSearch(w, r, queries, uid); err != nil {
		return
	} else {
		data.Blogs = blogs
		data.NoResults = emptyWords
		data.EmptyWords = noResults
	}

	handlers.TemplateHandler(w, r, "resultBlogsActionPage.gohtml", data)
}

func BlogSearch(w http.ResponseWriter, r *http.Request, queries *db.Queries, uid int32) ([]*db.Blog, bool, bool, error) {
	searchWords := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	var blogIds []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := queries.BlogsSearchFirst(r.Context(), sql.NullString{
				String: word,
				Valid:  true,
			})
			if err != nil {
				log.Printf("blogsSearchFirst Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return nil, false, false, err
			}
			blogIds = ids
		} else {
			ids, err := queries.BlogsSearchNext(r.Context(), db.BlogsSearchNextParams{
				Word: sql.NullString{
					String: word,
					Valid:  true,
				},
				Ids: blogIds,
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

	rows, err := queries.GetBlogEntriesByIdsDescending(r.Context(), blogIds)
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
