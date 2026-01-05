package forum

import (
	"net/http"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// CreateTopicPageForm holds the data for the create topic form.
type CreateTopicPageForm struct {
	Participants        string
	InvalidParticipants string
	Title               string
	Description         string
}

// CreateTopicPageWithPostTask renders the create topic page with a given post task.
func CreateTopicPageWithPostTask(w http.ResponseWriter, r *http.Request, postTask tasks.TaskString, formData *CreateTopicPageForm) {
	type Data struct {
		CreateTask tasks.TaskString
		FormData   *CreateTopicPageForm
	}
	data := Data{
		CreateTask: postTask,
		FormData:   formData,
	}
	handlers.TemplateHandler(w, r, ForumCreateTopicPageTmpl, data)
}
