package main

import (
	"log"
	"net/http"
)

func writingsArticleEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsArticleEditPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsArticleEditActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
		static void updateWriting(a4webcont &cont, int wid, int pwcid, char* title, char* abstract, char* body, int isitprivate, int language)
		{
			char *s1 = cont.sql.mysqlEscapeString(title);
			char *s2 = cont.sql.mysqlEscapeString(abstract);
			char *s3 = cont.sql.mysqlEscapeString(body);
			a4string query("UPDATE writing SET writingCategory_idwritingCategory=\"%d\", title=\"%s\", abstract=\"%s\", writting=\"%s\", private=\"%d\", language_idlanguage=\"%d\" WHERE idwriting=%d ",
							    				    pwcid,	     s1,	      s2,	   s3,      isitprivate,		     language,		    wid);
			free(s1);
			free(s2);
			free(s3);
			a4mysqlResult *result = cont.sql.query(query.raw());
			delete result;
			query.set("DELETE FROM writingSearch WHERE writing_idwriting=%d", wid);
			result = cont.sql.query(query.raw());
			delete result;
			addToGeneralSearch(cont, abstract, wid, "writingSearch", "writing_idwriting");
			addToGeneralSearch(cont, title, wid, "writingSearch", "writing_idwriting");
			addToGeneralSearch(cont, body, wid, "writingSearch", "writing_idwriting");
		}

			int pwcid = atoiornull(cont.post.getS("pwcid"));
		int wid = atoiornull(cont.post.getS("wid"));
		int language = atoiornull(cont.post.getS("language"));
		int isitprivate = cont.post.getS("isitprivate") != NULL;
		char *title = cont.post.getS("title");
		char *abstract = cont.post.getS("abstract");
		char *body = cont.post.getS("body");
		int isuserwriter = iswriter(cont, cont.user.UID, wid);
		int isabletoedit = isuserabletoeditwriting(cont, cont.user.UID, wid);
		if (level == auth_administrator || isuserwriter || isabletoedit)
			updateWriting(cont, wid, pwcid, title, abstract, body, isitprivate, language);

	*/
}
