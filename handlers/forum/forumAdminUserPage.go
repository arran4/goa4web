package forum

import (
	"database/sql"
	"errors"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/templates"
)

func AdminUserPage(w http.ResponseWriter, r *http.Request) {

	type UserTopic struct {
		User   *db.User
		Topics []*db.GetAllForumTopicsForUserWithPermissionsRestrictionsAndTopicRow
	}

	type Data struct {
		*CoreData
		Rows       []*UserTopic
		Categories map[int32]*db.Forumcategory
		Search     string
		NextLink   string
		PrevLink   string
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
		Search:   r.URL.Query().Get("search"),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

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
	data.Categories = make(map[int32]*db.Forumcategory)
	for _, c := range catRows {
		data.Categories[c.Idforumcategory] = c
	}

	if data.Search != "" {
		q := strings.ToLower(data.Search)
		var filtered []*db.User
		for _, u := range users {
			if strings.Contains(strings.ToLower(u.Username.String), q) ||
				strings.Contains(strings.ToLower(u.Email.String), q) {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	pageSize := common.GetPageSize(r)
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

	if err := templates.RenderTemplate(w, "adminUserPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
