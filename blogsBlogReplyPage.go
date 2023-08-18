package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func blogsBlogReplyPostPage(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Printf("Error: store.Get: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	blog, err := queries.show_blog(r.Context(), int32(bid))
	if err != nil {
		log.Printf("show_blog_comments Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	var pthid int32 = blog.ForumthreadIdforumthread
	ptid, err := queries.findForumTopicByName(r.Context(), sql.NullString{
		String: "A BLOGGER TOPIC",
		Valid:  true,
	})
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.makeTopic(r.Context(), makeTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: "A BLOGGER TOPIC",
				Valid:  true,
			},
			Description: sql.NullString{
				String: "THIS IS A HIDDEN FORUM FOR A BLOGGER TOPIC",
				Valid:  true,
			},
		})
		if err != nil {
			log.Printf("Error: makeTopic: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("Error: findForumTopicByName: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if pthid == 0 {
		pthidi, err := queries.makeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		pthid = int32(pthidi)
		if err := queries.assign_blog_to_thread(r.Context(), assign_blog_to_threadParams{
			ForumthreadIdforumthread: pthid,
			Idblogs:                  int32(bid),
		}); err != nil {
			log.Printf("Error: assign_blog_to_thread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/blogs/blog/%d/comments", bid)

	if rows, err := queries.threadNotify(r.Context(), threadNotifyParams{
		ForumthreadIdforumthread: pthid,
		Idusers:                  uid,
	}); err != nil {
		log.Printf("Error: threadNotify: %s", err)
	} else {
		for _, row := range rows {
			if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	if rows, err := queries.somethingNotifyBlogs(r.Context(), somethingNotifyBlogsParams{
		Idusers: uid,
		Idblogs: int32(bid),
	}); err != nil {
		log.Printf("Error: threadNotify: %s", err)
	} else {
		for _, row := range rows {
			if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)

			}
		}
	}

	if err := queries.makePost(r.Context(), makePostParams{
		LanguageIdlanguage:       int32(languageId),
		UsersIdusers:             uid,
		ForumthreadIdforumthread: pthid,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	}); err != nil {
		log.Printf("Error: makePost: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	/* TODO
	-- name: postUpdate :exec
	UPDATE comments c, forumthread th, forumtopic t
	SET
	th.lastposter=c.users_idusers, t.lastposter=c.users_idusers,
	th.lastaddition=c.written, t.lastaddition=c.written,
	t.comments=IF(th.comments IS NULL, 0, t.comments+1),
	t.threads=IF(th.comments IS NULL, IF(t.threads IS NULL, 1, t.threads+1), t.threads),
	th.comments=IF(th.comments IS NULL, 0, th.comments+1),
	th.firstpost=IF(th.firstpost=0, c.idcomments, th.firstpost)
	WHERE c.idcomments=?;
	*/
	if err := queries.update_forumthread(r.Context(), pthid); err != nil {
		log.Printf("Error: update_forumthread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.update_forumtopic(r.Context(), ptid); err != nil {
		log.Printf("Error: update_forumtopic: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)

}
