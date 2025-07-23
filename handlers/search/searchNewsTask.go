package search

import (
	"net/http"

	news "github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/tasks"
)

// SearchNewsTask performs a news search.
type SearchNewsTask struct{ tasks.TaskString }

var searchNewsTask = &SearchNewsTask{TaskString: TaskSearchNews}
var _ tasks.Task = (*SearchNewsTask)(nil)

func (SearchNewsTask) Action(w http.ResponseWriter, r *http.Request) any {
	return http.HandlerFunc(news.SearchResultNewsActionPage)
}
