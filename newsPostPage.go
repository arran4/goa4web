package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

type NewsPost struct {
	ShowReply bool
	ShowEdit  bool
	// TODO or (eq .Level "authWriter") (and (ge .Level "authModerator") (le .Level "authAdministrator"))
}

func newsPostPage(w http.ResponseWriter, r *http.Request) {
	type Post struct {
		*getNewsPostRow
		ShowReply bool
		ShowEdit  bool
		Editing   bool
	}
	type Data struct {
		*CoreData
		Post      *Post
		Languages []*Language
		Thread    *Forumthread
		Topic     *Forumtopic
		Comments  []*Comment
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	post, err := queries.getNewsPost(r.Context(), int32(pid))
	if err != nil {
		log.Printf("getNewsPost Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	editingId, _ := strconv.Atoi(r.URL.Query().Get("reply"))

	data.Post = &Post{
		getNewsPostRow: post,
		ShowReply:      true, // TODO
		ShowEdit:       true, // TODO
		Editing:        editingId == int(post.Idsitenews),
	}

	CustomNewsIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "newsPostPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func newsPostReplyActionPage(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Printf("Error: store.Get: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	pid, err := strconv.Atoi(vars["post"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if pid == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	post, err := queries.getNewsPost(r.Context(), int32(pid))
	if err != nil {
		log.Printf("getNewsPost Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pthid int32 = post.ForumthreadIdforumthread
	ptid, err := queries.findForumTopicByName(r.Context(), sql.NullString{
		String: "A NEWS TOPIC",
		Valid:  true,
	})
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.makeTopic(r.Context(), makeTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: "A NEWS TOPIC",
				Valid:  true,
			},
			Description: sql.NullString{
				String: "THIS IS A HIDDEN FORUM FOR A NEWS TOPIC",
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
		if err := queries.assignNewsThisThreadId(r.Context(), assignNewsThisThreadIdParams{
			ForumthreadIdforumthread: pthid,
			Idsitenews:               int32(pid),
		}); err != nil {
			log.Printf("Error: assign_news_to_thread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/news/%d", pid)

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

	// TODO
	//if rows, err := queries.somethingNotifyNews(r.Context(), somethingNotifyNewssParams{
	//	Idusers: uid,
	//	Idnewss: int32(bid),
	//}); err != nil {
	//	log.Printf("Error: threadNotify: %s", err)
	//} else {
	//	for _, row := range rows {
	//		if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
	//			log.Printf("Error: notifyChange: %s", err)
	//
	//		}
	//	}
	//}

	cid, err := queries.makePost(r.Context(), makePostParams{
		LanguageIdlanguage:       int32(languageId),
		UsersIdusers:             uid,
		ForumthreadIdforumthread: pthid,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
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

	wordIds, done := SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func newsPostEditActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO verify field names
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["post"])

	err = queries.editNewsPost(r.Context(), editNewsPostParams{
		Idsitenews:         int32(postId),
		LanguageIdlanguage: int32(languageId),
		News: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/news/%d", postId), http.StatusTemporaryRedirect)
}

func newsPostNewActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO verify field names
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["post"])
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	err = queries.writeNewsPost(r.Context(), writeNewsPostParams{
		LanguageIdlanguage: int32(languageId),
		News: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers: uid,
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/news/%d", postId), http.StatusTemporaryRedirect)
}
