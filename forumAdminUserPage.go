package goa4web

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func forumAdminUserPage(w http.ResponseWriter, r *http.Request) {

	type UserTopic struct {
		User   *User
		Topics []*GetAllForumTopicsForUserWithPermissionsRestrictionsAndTopicRow
	}

	type Data struct {
		*CoreData
		Rows       []*UserTopic
		Categories map[int32]*Forumcategory
		Search     string
		NextLink   string
		PrevLink   string
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Search:   r.URL.Query().Get("search"),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	users, err := queries.AllUsers(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	catRows, err := queries.GetAllForumCategories(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Categories = make(map[int32]*Forumcategory)
	for _, c := range catRows {
		data.Categories[c.Idforumcategory] = c
	}

	if data.Search != "" {
		q := strings.ToLower(data.Search)
		var filtered []*User
		for _, u := range users {
			if strings.Contains(strings.ToLower(u.Username.String), q) ||
				strings.Contains(strings.ToLower(u.Email.String), q) {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	pageSize := getPageSize(r)
	if offset < 0 {
		offset = 0
	}
	if offset > len(users) {
		offset = len(users)
	}
	end := offset + pageSize
	if end > len(users) {
		end = len(users)
	}

	for _, u := range users[offset:end] {
		topics, err := queries.GetAllForumTopicsForUserWithPermissionsRestrictionsAndTopic(r.Context(), u.Idusers)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getAllUsersTopicLevels Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data.Rows = append(data.Rows, &UserTopic{User: u, Topics: topics})
	}

	if data.Search != "" {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: fmt.Sprintf("Next %d", pageSize),
			Link: fmt.Sprintf("/forum/admin/users?search=%s&offset=%d", url.QueryEscape(data.Search), offset+pageSize),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: fmt.Sprintf("Previous %d", pageSize),
				Link: fmt.Sprintf("/forum/admin/users?search=%s&offset=%d", url.QueryEscape(data.Search), offset-pageSize),
			})
		}
	} else {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: fmt.Sprintf("Next %d", pageSize),
			Link: fmt.Sprintf("/forum/admin/users?offset=%d", offset+pageSize),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: fmt.Sprintf("Previous %d", pageSize),
				Link: fmt.Sprintf("/forum/admin/users?offset=%d", offset-pageSize),
			})
		}
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{Name: "Previous 15", Link: data.PrevLink})
	}

	CustomForumIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "forumAdminUserPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
