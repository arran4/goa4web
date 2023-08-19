package main

import (
	"log"
	"net/http"
)

func linkerAdminAddPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???
	CustomLinkerIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerAdminAddPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func linkerAdminAddActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
		static int addToLinker(a4webcont &cont, int userid, int langid, int catid, char *title, char* url, char *description)
		{
			char *t = cont.sql.mysqlEscapeString(title);
			char *u = cont.sql.mysqlEscapeString(url);
			char *d = cont.sql.mysqlEscapeString(description);
			a4string query("INSERT INTO linker (users_idusers, linkerCategory_idlinkerCategory, title, url,   description, listed) VALUES "
							  "(%d,            %d,                              \"%s\",\"%s\",\"%s\",      NOW() );",
							  userid,          catid,                           t,     u,     d);
			a4mysqlResult *result = cont.sql.query(query.raw());
			free(t);
			free(u);
			free(d);
			int value = cont.sql.last_insert_id();
			delete result;
			addToGeneralSearch(cont, description, value, "linkerSearch", "linker_idlinker");
			addToGeneralSearch(cont, title, value, "linkerSearch", "linker_idlinker");
			return value;
		}

	*/
}
