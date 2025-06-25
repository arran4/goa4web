package goa4web

import (
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func forumAdminThreadsPage(w http.ResponseWriter, r *http.Request) {
	type Group struct {
		TopicTitle string
		Threads    []*GetAllForumThreadsWithTopicRow
	}
	type Data struct {
		*CoreData
		Groups map[int32]*Group
		Order  []int32
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.GetAllForumThreadsWithTopic(r.Context())
	if err != nil {
		log.Printf("GetAllForumThreadsWithTopic: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Groups:   make(map[int32]*Group),
	}

	for _, row := range rows {
		g, ok := data.Groups[row.ForumtopicIdforumtopic]
		if !ok {
			g = &Group{TopicTitle: row.TopicTitle.String}
			data.Groups[row.ForumtopicIdforumtopic] = g
			data.Order = append(data.Order, row.ForumtopicIdforumtopic)
		}
		g.Threads = append(g.Threads, row)
	}

	CustomForumIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "adminThreadsPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func forumAdminThreadDeletePage(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(mux.Vars(r)["thread"])
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	topicID, err := strconv.Atoi(r.PostFormValue("topic"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	if err := ThreadDelete(r.Context(), queries, int32(threadID), int32(topicID)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/forum/admin/conversations", http.StatusTemporaryRedirect)
}
