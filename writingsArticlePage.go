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

func writingsArticlePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Writing   *fetchWritingByIdRow
		CanEdit   bool
		IsAuthor  bool
		CanReply  bool
		UserId    int32
		Languages []*Language
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		CanReply: true,  // TODO
		CanEdit:  false, // TODO
	}

	vars := mux.Vars(r)
	articleId, _ := strconv.Atoi(vars["article"])

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	writing, err := queries.fetchWritingById(r.Context(), fetchWritingByIdParams{
		Userid:    uid,
		Idwriting: int32(articleId),
	})
	if err != nil {
		log.Printf("fetchWritingById Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Writing = writing
	data.IsAuthor = writing.UsersIdusers == uid

	languageRows, err := queries.fetchLanguages(r.Context())
	if err != nil {
		log.Printf("fetchLanguages Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomWritingsIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsArticlePage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsArticleReplyActionPage(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	vars := mux.Vars(r)
	aid, err := strconv.Atoi(vars["post"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if aid == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	uid, _ := session.Values["UID"].(int32)

	post, err := queries.fetchWritingById(r.Context(), fetchWritingByIdParams{
		Userid:    uid,
		Idwriting: int32(aid),
	})
	if err != nil {
		log.Printf("getArticlePost Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pthid int32 = post.ForumthreadIdforumthread
	ptid, err := queries.findForumTopicByName(r.Context(), sql.NullString{
		String: WritingTopicName,
		Valid:  true,
	})
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.makeTopic(r.Context(), makeTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: WritingTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: WritingTopicDescription,
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
		if err := queries.assignWritingThisThreadId(r.Context(), assignWritingThisThreadIdParams{
			ForumthreadIdforumthread: pthid,
			Idwriting:                int32(aid),
		}); err != nil {
			log.Printf("Error: assign_article_to_thread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	endUrl := fmt.Sprintf("/article/%d", aid)

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
	//if rows, err := queries.somethingNotifyArticle(r.Context(), somethingNotifyArticlesParams{
	//	Idusers: uid,
	//	Idarticles: int32(bid),
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

	taskDoneAutoRefreshPage(w, r)
	/*
		TODO
			{
			int ptid = atoiornull(cont.post.getS("replyTo"));
			if (!article)
				return;
			if (!ptid)
			{
				ptid = findForumTopicByName(cont, WRITING_TOPIC_NAME);
				if (ptid == 0)
					ptid = makeTopic(cont, 0, WRITING_TOPIC_NAME,WRITING_TOPIC_DESCRIPTION);
				if (!ptid)
				{
					printf("Failed to create invisible topic.<br>\n");
					return;
				}
				ptid = makeThread(cont, ptid);
				if (!ptid)
				{
					printf("Failed to create thread.<br>\n");
					return;
				}
				assignWritingThisThreadId(cont, ptid, article);
			}
			int langid = atoiornull(cont.post.getS("language"));
			int boardid = atoiornull((char*)cont.get.get("board"));
			a4string querystr("section=read&article=%d&category=%d", article, category);
			threadNotify(cont, ptid, querystr.raw());
			somethingNotify(cont, "writing", "idwriting", article, querystr.raw(), "users_idusers");
			int lastinsert = makePost(cont, ptid, cont.post.getS("replytext"), langid);
			if (!lastinsert)
			{
				printf("Failed to create post.<br>\n");
				return;
			}


	*/
}
