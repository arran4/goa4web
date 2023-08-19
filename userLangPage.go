package main

import (
	"log"
	"net/http"
)

func userLangPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	if err := getCompiledTemplates().ExecuteTemplate(w, "userLangPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func userLangSaveLanguagesActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*

			if(!strcasecmp((char*)cont.post.getS("dothis"), "Save languages") || !strcasecmp((char*)cont.post.getS("dothis"), "Save all"))
		{
			a4string querystr("DELETE FROM userlang WHERE users_idusers=\"%d\"", cont.user.UID);
			a4mysqlResult *result = cont.sql.query(querystr.raw());
			delete result;
			querystr.set("SELECT idlanguage FROM language");
			result = cont.sql.query(querystr.raw());
			querystr.set("INSERT INTO userlang (users_idusers, language_idlanguage) VALUES ");
			int addcount = 0;
			while (result->hasRow())
			{
				a4string tmp("language%s", result->getColumn(0));
				if (cont.post.getS(tmp.raw()) != NULL)
					querystr.pushf("%s(\"%d\", \"%s\")", addcount++ ? "," : "", cont.user.UID, result->getColumn(0));
				result->nextRow();
			}
			delete result;
			if (addcount)
			{
				result = cont.sql.query(querystr.raw());
				delete result;
			}
		}
		if (!strcasecmp((char*)cont.post.getS("dothis"), "Save language") || !strcasecmp((char*)cont.post.getS("dothis"), "Save all"))
		{
			a4string querystr("SELECT COUNT(users_idusers) FROM preferences WHERE users_idusers=\"%d\"", cont.user.UID);
			a4mysqlResult *result = cont.sql.query(querystr.raw());
			int prefcount = atoiornull(result->getColumn(0));
			delete result;
			if (prefcount)
			{
				a4string querystr("UPDATE preferences SET language_idlanguage=\"%d\" WHERE users_idusers=\"%d\"",
						atoiornull((char*)cont.post.getS("language")), cont.user.UID);
				a4mysqlResult *result = cont.sql.query(querystr.raw());
				delete result;
			} else {
				a4string querystr("INSERT INTO preferences (language_idlanguage, users_idusers) VALUES (\"%d\", \"%d\")",
						atoiornull((char*)cont.post.getS("language")), cont.user.UID);
				a4mysqlResult *result = cont.sql.query(querystr.raw());
				delete result;
			}
		}


	*/
}
