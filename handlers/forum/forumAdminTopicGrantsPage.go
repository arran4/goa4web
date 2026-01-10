package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// AdminTopicGrantsPage displays the drag-and-drop role grants editor for a forum topic.
func AdminTopicGrantsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		TopicID     int32
		UpdateURL   string
		GrantGroups []TopicGrantGroup
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum - Topic %d Grants", tid)
	groups, err := buildTopicGrantGroups(r.Context(), cd, int32(tid))
	if err != nil {
		log.Printf("buildTopicGrantGroups: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data := Data{
		TopicID:     int32(tid),
		UpdateURL:   strings.TrimSuffix(r.URL.Path, "/grants") + "/grant/update",
		GrantGroups: groups,
	}
	ForumAdminTopicGrantsPageTmpl.Handle(w, r, data)
}

const ForumAdminTopicGrantsPageTmpl handlers.Page = "forum/adminTopicGrantsPage.gohtml"
