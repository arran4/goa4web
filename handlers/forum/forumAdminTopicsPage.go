package forum

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminTopicsPage shows all forum topics for management.
func AdminTopicsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Topics     []*db.Forumtopic
		Categories []*db.Forumcategory
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	topics, err := queries.GetAllForumTopics(r.Context())
	if err != nil {
		log.Printf("GetAllForumTopics: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	cats, err := queries.GetAllForumCategories(r.Context())
	if err != nil {
		log.Printf("GetAllForumCategories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData:   r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Topics:     topics,
		Categories: cats,
	}

	handlers.TemplateHandler(w, r, "adminTopicsPage.gohtml", data)
}
