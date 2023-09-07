package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func CustomNewsIndex(data *CoreData, r *http.Request) {
	// TODO
	// TODO RSS
	userHasAdmin := true // TODO
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "User Permissions",
			Link: "/news/user/permissions",
		})
	}
	userHasWriter := true // TODO
	if userHasWriter || userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Add News",
			Link: "/news/post",
		})
	}

	vars := mux.Vars(r)
	newsId, _ := vars["news"]
	if newsId != "" {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Return to list",
			Link: fmt.Sprintf("/?offset=%d", 0),
		})
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset != 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "The start",
			Link: fmt.Sprintf("?offset=%d", 0),
		})
	}
	data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
		Name: "Next 10",
		Link: fmt.Sprintf("?offset=%d", offset+10),
	})
	if offset > 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Previous 10",
			Link: fmt.Sprintf("?offset=%d", offset-10),
		})
	}
}
