package main

import (
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func searchResultWritingsActionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	vars := mux.Vars(r)

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	/* TODO General search
		int count = 0;
	while (*keys != NULL)
	{
		char *s1 = cont.sql.mysqlEscapeString(*keys);
		query.set("SELECT st.%s_id%s FROM searchwordlist swl, %s st WHERE swl.word=\"%s\" AND "
				"swl.idsearchwordlist=st.searchwordlist_idsearchwordlist" , table, table, searchtable, s1);
		if (count++)
		{
			query.pushf(" AND cs.comments_idcomments IN (%s)", inlist.raw());
		}
		free(s1);
		result = cont.sql.query(query.raw(), false);
		inlist.clear();
		int i = 0;
		while (result->hasRow())
		{
			i++;
			inlist.pushf("%s", result->getColumn(0));
			if (result->nextRow())
				inlist.pushf(", ");
		}
		delete result;
		if (i == 0)
		{
			printf("Nothing found.<br>\n");
			return;
		}
		keys++;
	}

	*/

	/*

		comment search

			forumTopicSearch(cont, searchwords, "A WRITING TOPIC", "w.idwriting",
				"LEFT JOIN writing w ON th.idforumthread=w.forumthread_idforumthread", "writings.cgi?section=read&article=");

		int topicid = findForumTopicByName(cont, topicName);
		a4string query, inlist;
		a4mysqlResult *result;
		a4hashtable words;
		a4hashtable nowords;
		//nowords.set("this", (void*)1);
		breakupTextToWords(cont, sw, words, nowords);
		char **keys = words.keys();
		int count = 0;
		while (*keys != NULL)
		{
			char *s1 = cont.sql.mysqlEscapeString(*keys);
			query.set("SELECT cs.comments_idcomments FROM searchwordlist swl, commentsSearch cs WHERE swl.word=\"%s\" AND "
					"swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist" , s1);
			if (count++)
			{
				query.pushf(" AND cs.comments_idcomments IN (%s)", inlist.raw());
			}
			free(s1);
			result = cont.sql.query(query.raw(), false);
			inlist.clear();
			int i = 0;
			while (result->hasRow())
			{
				i++;
				inlist.pushf("%s", result->getColumn(0));
				if (result->nextRow())
					inlist.pushf(", ");
			}
			delete result;
			if (i == 0)
			{
				printf("Nothing found.<br>\n");
				return;
			}
			keys++;
		}

	*/

	// Custom Index???

	if err := getCompiledTemplates().ExecuteTemplate(w, "searchResultWritingsActionPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
