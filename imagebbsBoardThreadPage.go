package main

import (
	"log"
	"net/http"
)

func imagebbsBoardThreadPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		ImagePosts []*ImagePosts
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomImageBBSIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "imagebbsBoardThreadPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func imagebbsBoardThreadReplyActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO

	ptid = findForumTopicByName(cont, "A IMAGE BBS TOPIC");
	if (ptid == 0)
		ptid = makeTopic(cont, 0, "A IMAGE BBS TOPIC","THIS IS A HIDDEN FORUM FOR A IMAGE BBS");
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

	queries.assignImagePostThisThreadId

	threadNotify(cont, ptid, querystr.raw());
	somethingNotify(cont, "imagepost", "idimagepost", ipid, querystr.raw(), "users_idusers");
	int lastinsert = makePost(cont, ptid, cont.post.getS("replytext"), langid);
	updateSearch

	taskDoneAutoRefreshPage
}
