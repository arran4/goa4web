package admin

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

type grantRuleTopic struct {
	ID    int32
	Title string
}

// AdminGrantRulesPage lists forum topics that only have user grants.
func AdminGrantRulesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		IncludeAdmin bool
		Topics       []*grantRuleTopic
		TaskName     string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Grant Config Rules"
	includeAdmin := r.URL.Query().Get("include_admin") == "1"
	rows, err := cd.Queries().AdminListTopicsWithUserGrantsNoRoles(r.Context(), includeAdmin)
	if err != nil {
		log.Printf("list topics with user grants: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	topics := make([]*grantRuleTopic, 0, len(rows))
	for _, row := range rows {
		title := ""
		if row.Title.Valid {
			title = row.Title.String
		}
		topics = append(topics, &grantRuleTopic{ID: row.Idforumtopic, Title: title})
	}
	data := Data{IncludeAdmin: includeAdmin, Topics: topics, TaskName: string(TaskForumTopicConvertPrivate)}
	handlers.TemplateHandler(w, r, "grantRulesPage.gohtml", data)
}
