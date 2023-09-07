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

func linkerShowPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Link               *ShowLinkRow
		CanReply           bool
		Languages          []*Language
		SelectedLanguageId int
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		CanReply: true, // TODO
	}

	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	link, err := queries.ShowLink(r.Context(), int32(linkId))
	if err != nil {
		log.Printf("showLink Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Link = link

	CustomLinkerIndex(data.CoreData, r)
	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerShowPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerShowReplyPage(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	vars := mux.Vars(r)
	linkId, err := strconv.Atoi(vars["link"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if linkId == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	link, err := queries.ShowLink(r.Context(), int32(linkId))
	if err != nil {
		log.Printf("showLink Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pthid int32 = link.ForumthreadIdforumthread
	ptid, err := queries.FindForumTopicByName(r.Context(), sql.NullString{
		String: LinkderTopicName,
		Valid:  true,
	})
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.MakeTopic(r.Context(), MakeTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: LinkderTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: LinkderTopicName,
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
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		pthid = int32(pthidi)
		if err := queries.AssignLinkerThisThreadId(r.Context(), AssignLinkerThisThreadIdParams{
			ForumthreadIdforumthread: pthid,
			Idlinker:                 int32(linkId),
		}); err != nil {
			log.Printf("Error: assign_blog_to_thread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/linker/show/%d", linkId)

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), ListUsersSubscribedToThreadParams{
		ForumthreadIdforumthread: pthid,
		Idusers:                  uid,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else {
		for _, row := range rows {
			if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	if rows, err := queries.ListUsersSubscribedToLinker(r.Context(), ListUsersSubscribedToLinkerParams{
		Idusers:  uid,
		Idlinker: int32(linkId),
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else {
		for _, row := range rows {
			if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)

			}
		}
	}

	cid, err := queries.MakePost(r.Context(), MakePostParams{
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
	if err := queries.Update_forumthread(r.Context(), pthid); err != nil {
		log.Printf("Error: update_forumthread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.Update_forumtopic(r.Context(), ptid); err != nil {
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
