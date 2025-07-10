package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/emailutil"
	searchutil "github.com/arran4/goa4web/internal/searchutil"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

func BlogReplyPostPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	if err := hcommon.ValidateForm(r, []string{"language", "replytext"}, []string{"language", "replytext"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	bid, err := strconv.Atoi(vars["blog"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if bid == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)

	blog, err := queries.GetBlogEntryForUserById(r.Context(), int32(bid))
	if err != nil {
		log.Printf("getBlogEntryForUserById_comments Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	var pthid int32
	if blog.ForumthreadID.Valid {
		pthid = blog.ForumthreadID.Int32
	}
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: BloggerTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: BloggerTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: BloggerTopicDescription,
				Valid:  true,
			},
		})
		if err != nil {
			log.Printf("Error: createForumTopic: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("Error: findForumTopicByTitle: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	} else {
		ptid = pt.Idforumtopic
	}
	if pthid == 0 {
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		pthid = int32(pthidi)
		if err := queries.AssignThreadIdToBlogEntry(r.Context(), db.AssignThreadIdToBlogEntryParams{
			ForumthreadID: sql.NullInt32{Int32: pthid, Valid: true},
			Idblogs:       int32(bid),
		}); err != nil {
			log.Printf("Error: assignThreadIdToBlogEntry: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/blogs/blog/%d/comments", bid)

	provider := email.ProviderFromConfig(config.AppRuntimeConfig)

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), db.ListUsersSubscribedToThreadParams{
		ForumthreadID: pthid,
		Idusers:       uid,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, row.Idusers, row.Email, endUrl, "update", nil); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	if rows, err := queries.ListUsersSubscribedToBlogs(r.Context(), db.ListUsersSubscribedToBlogsParams{
		Idusers: uid,
		Idblogs: int32(bid),
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, row.Idusers, row.Email, endUrl, "update", nil); err != nil {
				log.Printf("Error: notifyChange: %s", err)

			}
		}
	}

	cid, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
		ForumthreadID:      pthid,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		log.Printf("Error: createComment: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := PostUpdate(r.Context(), queries, pthid, ptid); err != nil {
		log.Printf("Error: postUpdate: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	wordIds, done := searchutil.SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if searchutil.InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)

}
