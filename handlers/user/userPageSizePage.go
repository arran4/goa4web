package user

import (
	"net/http"
	"strconv"

	handlers "github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/config"
)

type PageSizeSaveTask struct{ tasks.TaskString }

var pageSizeSaveTask = &PageSizeSaveTask{TaskString: TaskSaveAll}

func userPageSizePage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			min, _ := strconv.Atoi(r.PostFormValue("min"))
			max, _ := strconv.Atoi(r.PostFormValue("max"))
			def, _ := strconv.Atoi(r.PostFormValue("default"))
			config.UpdatePaginationConfig(&config.AppRuntimeConfig, min, max, def)
		}
		http.Redirect(w, r, "/usr/page-size", http.StatusSeeOther)
		return
	}
	data := struct {
		*handlers.CoreData
		Min     int
		Max     int
		Default int
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
		Min:      config.AppRuntimeConfig.PageSizeMin,
		Max:      config.AppRuntimeConfig.PageSizeMax,
		Default:  config.AppRuntimeConfig.PageSizeDefault,
	}
	handlers.TemplateHandler(w, r, "pageSizePage.gohtml", data)
}

func (PageSizeSaveTask) Action(w http.ResponseWriter, r *http.Request) {
	userPageSizePage(w, r)
}
