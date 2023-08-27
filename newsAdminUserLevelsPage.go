package main

import (
	"log"
	"net/http"
)

func newsAdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// SKIP. TODO replace completely
	// Custom Index???

	if err := getCompiledTemplates().ExecuteTemplate(w, "newsAdminUserLevelsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func newsAdminUserLevelsAllowActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
			char *postusername = cont.post.getS("username");
		if (postusername == NULL)
			postusername = "0";
		int userid = usernametouid(cont, postusername);
		char *postlevel = cont.post.getS("level");
		if (postlevel == NULL)
			postlevel = "0";
		s = cont.sql.mysqlEscapeString(postlevel);
		if (userid && s != NULL)
		{
			a4string query("INSERT INTO permissions "
					"(users_idusers, section, level)"
					"VALUES (\"%d\",\"%s\",\"%s\")",
					userid, pageName, s);
			a4mysqlResult *result = cont.sql.query(query.raw());
			if (cont.sql.errno())
				printf("Error with query: %s<br>\n",
						cont.sql.error());
			delete result;
		} else {
			if (s == NULL)
				printf("Error with encoding string.<br>\n");
			if (userid == 0)
				printf("Error with username %s<br>\n",
						cont.sql.error());
		}

	*/
	taskDoneAutoRefreshPage(w, r)
}

func newsAdminUserLevelsRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
			char *postid = cont.post.getS("id");
		if (postid == NULL)
			postid = "0";
		a4string query("DELETE FROM permissions WHERE idpermissions=%d",
				atoi(postid));
		a4mysqlResult *result = cont.sql.query(query.raw());
		delete result;

	*/
	taskDoneAutoRefreshPage(w, r)
}
