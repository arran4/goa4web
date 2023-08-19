package main

import (
	"log"
	"net/http"
)

type NewsPost struct {
	ShowReply bool
	ShowEdit  bool
	// TODO or (eq .Level "authWriter") (and (ge .Level "authModerator") (le .Level "authAdministrator"))
}

func newsPostPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Post      *NewsPost
		Languages []*Language
		Thread    *Forumthread
		Topic     *Forumtopic
		Comments  []*Comment
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomNewsIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "newsPostPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func newsPostReplyActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
		if (dowhat != NULL && !strcasecmp("Reply", dowhat))
		{
			if (cont.user.UID == 0) return;
			int ptid = getNewsThreadId(cont, newsid);
			if (!ptid)
			{
				ptid = findForumTopicByName(cont, "A NEWS TOPIC");
				if (ptid == 0)
					ptid = makeTopic(cont, 0, "A NEWS TOPIC","THIS IS A HIDDEN FORUM FOR A NEWS TOPIC");
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
				assignNewsThisThreadId(cont, ptid, newsid);
			}
			langid = atoiornull(cont.post.getS("language"));
			int siteNewsid = atoiornull((char*)cont.get.get("show"));
			a4string querystr("show=%d", siteNewsid);
			threadNotify(cont, ptid, querystr.raw());
			somethingNotify(cont, "siteNews", "idsiteNews", siteNewsid, querystr.raw(), "users_idusers");
			int lastinsert = makePost(cont, ptid, cont.post.getS("replytext"), langid);
			if (!lastinsert)
			{
				printf("Failed to create post.<br>\n");
				return;
			}
		}

	*/
}
func newsPostEditActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
		char *idstr = (char*)cont.get.get("id");
		if (idstr == NULL) return;
		int id = atoi(idstr);
		a4string query("SELECT s.news, s.idsiteNews, u.idusers, s.language_idlanguage "
				"FROM siteNews s, users u "
				"WHERE s.users_idusers=u.idusers AND s.idsiteNews=\"%d\"",
				id);
		a4mysqlResult *result = cont.sql.query(query.raw());
		if (level == auth_writer && atoi(result->getColumn(2)) != cont.user.UID) return;
		int stage = cont.post.getS("exec") == NULL;
		if (stage)
		{
			printf("<form method=post>"
				"<input type=hidden name=\"doto\" value=\"%d\">"
				"<textarea name=\"newstext\" cols=40 rows=20>%s</textarea><br>",
				id, result->getColumn(0));
			int lang = atoi(result->getColumn(3));
			delete result;
			lang_combobox(cont, lang);
			printf("<input type=submit name=\"exec\" value=\"Edit\">"
				"</form>");
		} else delete result;
		if (stage == 0)
		{
			printf("Editing post.<br>\n");
			int langid = atoiornull(cont.post.getS("language"));
			editNewsPost(cont, id, langid, cont.post.getS("newstext"));
			printf("Post Edited.<br>\n");
		}
		clear search for entry then re-add -- find diff?
		addToGeneralSearch(cont, text, postid, "siteNewsSearch", "siteNews_idsiteNews");
	*/
	taskDoneAutoRefreshPage(w, r)
}
func newsPostNewActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
		if (cont.post.getS("exec") != NULL)
		{
			printf("Writing post.<br>\n");
			int langid = atoiornull(cont.post.getS("language"));
			writeNewsPost(cont, langid, cont.post.getS("newstext"));
			printf("Post written.<br>\n");
		} else
		{
			printf("<form method=post>"
				"<textarea name=\"newstext\" cols=40 rows=20></textarea><br>");
			lang_combobox(cont, cont.pref.defaultLang);
			printf("<input type=submit name=\"exec\" value=\"Add\">"
					"</form>");
		}
		addToGeneralSearch(cont, text, lastinsert, "siteNewsSearch", "siteNews_idsiteNews");
	*/
	taskDoneAutoRefreshPage(w, r)
}
