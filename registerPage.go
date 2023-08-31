package main

import (
	"log"
	"net/http"
)

func registerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	if err := getCompiledTemplates().ExecuteTemplate(w, "registerPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func registerActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO

	/*

			switch(id = cont.user.registerUser(username, password, email))
		{
			case -1:
				begin_page(cont);
				printf("User already exists.");
				break;
			case -2:
				begin_page(cont);
				printf("Malformed data.");
				break;
			default:
				cont.user.setUser(id);
				begin_page(cont);
				printf("Success, user registered. You are number %d\n", id);
		}

	*/
}
