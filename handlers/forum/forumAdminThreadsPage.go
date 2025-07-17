package forum

import (
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
)

func AdminThreadsPage(w http.ResponseWriter, r *http.Request) {
	type Group struct {
		TopicTitle string
		Threads    []*db.GetAllForumThreadsWithTopicRow
	}
	type Data struct {
		*CoreData
		Groups map[int32]*Group
		Order  []int32
	}

	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

	rows, err := queries.GetAllForumThreadsWithTopic(r.Context())
	if err != nil {
		log.Printf("GetAllForumThreadsWithTopic: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*CoreData),
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

	common.TemplateHandler(w, r, "adminThreadsPage.gohtml", data)
}

func AdminThreadDeletePage(w http.ResponseWriter, r *http.Request) {
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
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	if err := ThreadDelete(r.Context(), queries, int32(threadID), int32(topicID)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/forum/admin/conversations", http.StatusTemporaryRedirect)
}
