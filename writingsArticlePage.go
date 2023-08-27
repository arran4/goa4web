package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func writingsArticlePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	vars := mux.Vars(r)

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	CustomWritingsIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsArticlePage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsArticleReplyActionPage(w http.ResponseWriter, r *http.Request) {
	/*
		{
		int ptid = atoiornull(cont.post.getS("replyTo"));
		if (!article)
			return;
		if (!ptid)
		{
			ptid = findForumTopicByName(cont, "A WRITING TOPIC");
			if (ptid == 0)
				ptid = makeTopic(cont, 0, "A WRITING TOPIC","THIS IS A HIDDEN FORUM FOR A WRITING");
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
