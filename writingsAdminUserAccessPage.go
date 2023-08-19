package main

import (
	"log"
	"net/http"
)

func writingsAdminUserAccessPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsAdminUserAccessPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsAdminUserAccessAllowActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*

		int wid = atoiornull(cont.post.getS("wid"));
		//int uid = atoiornull(cont.post.getS("uid"));
		char *username = cont.post.getS("username");
		if (username == NULL)
			return;
		int uid = usernametouid(cont, username);
		if (!uid)
			return;
		int readdoc = cont.post.getS("readdoc") != NULL;
		int editdoc = cont.post.getS("editdoc") != NULL;
		int isuserwriter = iswriter(cont, cont.user.UID, wid);
		if (level == auth_administrator || isuserwriter)
			addApprovedUser(cont, wid, uid, readdoc, editdoc);

	*/
}

func writingsAdminUserUpdateAllowActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*

			int wid = atoiornull(cont.post.getS("wid"));
		int uid = atoiornull(cont.post.getS("uid"));
		if (!uid)
			return;
		int readdoc = cont.post.getS("readdoc") != NULL;
		int editdoc = cont.post.getS("editdoc") != NULL;
		int isuserwriter = iswriter(cont, cont.user.UID, wid);
		if (level == auth_administrator || isuserwriter)
			setApprovedUser(cont, wid, uid, readdoc, editdoc);

	*/
}

func writingsAdminUserAccessRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO

	/*

		{
		int wid = atoiornull(cont.post.getS("wid"));
		int uid = atoiornull(cont.post.getS("uid"));
		int isuserwriter = iswriter(cont, cont.user.UID, wid);
		if (level == auth_administrator || isuserwriter)
			deleteApprovedUser(cont, wid, uid);

	*/
}
