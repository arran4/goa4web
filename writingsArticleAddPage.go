package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func writingsArticleAddPage(w http.ResponseWriter, r *http.Request) {
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

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsArticleAddPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func writingsArticleAddActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO

	/*

		static int makeWriting(a4webcont &cont, int pwcid, char* title, char* abstract, char* body, int isitprivate, int language)
		{
			char *s1 = cont.sql.mysqlEscapeString(title);
			char *s2 = cont.sql.mysqlEscapeString(abstract);
			char *s3 = cont.sql.mysqlEscapeString(body);
			a4string query("INSERT INTO writing (writingCategory_idwritingCategory, title, abstract, writting, private, language_idlanguage, published, users_idusers) VALUES "
					"(		    \"%d\", 				\"%s\", \"%s\", \"%s\", \"%d\", \"%d\", 	     NOW(),	\"%d\")",
							    pwcid, 				s1,     s2,      s3, isitprivate, language,			cont.user.UID);
			free(s1);
			free(s2);
			free(s3);
			a4mysqlResult *result = cont.sql.query(query.raw());
			int value = cont.sql.last_insert_id();
			delete result;
			addToGeneralSearch(cont, abstract, value, "writingSearch", "writing_idwriting");
			addToGeneralSearch(cont, title, value, "writingSearch", "writing_idwriting");
			addToGeneralSearch(cont, body, value, "writingSearch", "writing_idwriting");
			return value;
		}




			int pwcid = atoiornull(cont.post.getS("pwcid"));
		int language = atoiornull(cont.post.getS("language"));
		int isitprivate = cont.post.getS("isitprivate") != NULL;
		char *title = cont.post.getS("title");
		char *abstract = cont.post.getS("abstract");
		char *body = cont.post.getS("body");
		makeWriting(cont, pwcid, title, abstract, body, isitprivate, language);


	*/
}
