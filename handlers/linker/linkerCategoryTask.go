package linker

import (
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type linkerCategoryTask struct {
}

const (
	LinkerCategoryPageTmpl = "linker/categoryPage.gohtml"
)

func NewLinkerCategoryTask() tasks.Task {
	return &linkerCategoryTask{}
}

func (t *linkerCategoryTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{LinkerCategoryPageTmpl}
}

func (t *linkerCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *linkerCategoryTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	var data struct {
		Offset      int32
		HasOffset   bool
		CatId       int32
		CommentOnId int
		ReplyToId   int
	}

	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		data.Offset = int32(off)
	}
	data.HasOffset = data.Offset != 0
	vars := mux.Vars(r)
	if cid, err := strconv.Atoi(vars["category"]); err == nil {
		data.CatId = int32(cid)
	}
	if cid, err := strconv.Atoi(r.URL.Query().Get("comment")); err == nil {
		data.CommentOnId = cid
	}
	if rid, err := strconv.Atoi(r.URL.Query().Get("reply")); err == nil {
		data.ReplyToId = rid
	}

	if cat, err := cd.SelectedLinkerCategory(data.CatId); err == nil && cat != nil && cat.Title.Valid {
		cd.PageTitle = fmt.Sprintf("Category: %s", cat.Title.String)
	} else {
		cd.PageTitle = fmt.Sprintf("Category %d", data.CatId)
	}
	if err := cd.ExecuteSiteTemplate(w, r, LinkerCategoryPageTmpl, data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
