package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func linkerCommentsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	vars := mux.Vars(r)

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	// Custom Index???

	CustomLinkerIndex(data.CoreData, r)
	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerCommentsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func linkerCommentsReplyPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	// TODO
	/* 		int ptid = atoiornull(cont.post.getS("replyTo"));
	int lpid = atoiornull(cont.post.getS("lpid"));
	if (!ptid)
	{
		ptid = findForumTopicByName(cont, "A LINKER TOPIC");
		if (ptid == 0)
			ptid = makeTopic(cont, 0, "A LINKER TOPIC","THIS IS A HIDDEN FORUM FOR A LINKER TOPIC");
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
		assignLinkerThisThreadId(cont, ptid, lpid);
	}
	int langid = atoiornull(cont.post.getS("language"));
	a4string querystr("section=comment&link=%d", lpid);
	threadNotify(cont, ptid, querystr.raw());
	somethingNotify(cont, "linker", "idlinker", lpid, querystr.raw(), "users_idusers");
	int lastinsert = makePost(cont, ptid, cont.post.getS("replytext"), langid);
	if (!lastinsert)
	{
		printf("Failed to create post.<br>\n");
		return;
	}
	*/
}
