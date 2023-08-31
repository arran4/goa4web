package main

import (
	"log"
	"net/http"
)

func loginPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	if err := getCompiledTemplates().ExecuteTemplate(w, "loginPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func loginActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*

			switch(id = cont.user.login(username, password))
		{
			case -1:
				begin_page(cont);
				printf("Username of password incorrect.");
				break;
			case -2:
				begin_page(cont);
				printf("Malformed data.");
				break;
			case -3:
				begin_page(cont);
				printf("Database error.");
				break;
			default:
				cont.user.setUser(id);
				begin_page(cont);
				printf("Success Logged in.\n", id);
		}

	*/
}
