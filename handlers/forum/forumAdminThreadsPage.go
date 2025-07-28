package forum

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
)

func AdminThreadsPage(w http.ResponseWriter, r *http.Request) {
	type Group struct {
		TopicTitle string
		Threads    []*db.GetAllForumThreadsWithTopicRow
	}
	type Data struct {
		*common.CoreData
		Groups map[int32]*Group
		Order  []int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Threads"
	queries := cd.Queries()

	rows, err := queries.GetAllForumThreadsWithTopic(r.Context())
	if err != nil {
		log.Printf("GetAllForumThreadsWithTopic: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData: cd,
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

	handlers.TemplateHandler(w, r, "adminThreadsPage.gohtml", data)
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
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := ThreadDelete(r.Context(), queries, int32(threadID), int32(topicID)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/admin/forum/conversations", http.StatusTemporaryRedirect)
}
