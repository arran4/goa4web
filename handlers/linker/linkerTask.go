package linker

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type linkerTask struct {
}

const (
	LinkerPageTmpl handlers.Page = "linker/page.gohtml"
)

func NewLinkerTask() tasks.Task {
	return &linkerTask{}
}

func (t *linkerTask) TemplatesRequired() []string {
	return []string{string(LinkerPageTmpl)}
}

func (t *linkerTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *linkerTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cd.PageTitle = "Links"
	type Data struct {
		Offset      int32
		HasOffset   bool
		CatId       int32
		CommentOnId int
		ReplyToId   int
	}

	data := Data{}
	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		data.Offset = int32(off)
	}
	data.HasOffset = data.Offset != 0
	if cid, err := strconv.Atoi(r.URL.Query().Get("category")); err == nil {
		data.CatId = int32(cid)
	}
	if cid, err := strconv.Atoi(r.URL.Query().Get("comment")); err == nil {
		data.CommentOnId = cid
	}
	if rid, err := strconv.Atoi(r.URL.Query().Get("reply")); err == nil {
		data.ReplyToId = rid
	}

	offset := int(data.Offset)
	ps := cd.PageSize()
	vars := mux.Vars(r)
	categoryID := vars["category"]
	base := "/linker"
	if categoryID != "" {
		base = fmt.Sprintf("/linker/category/%s", categoryID)
	}
	cd.NextLink = fmt.Sprintf("%s?offset=%d", base, offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("%s?offset=%d", base, offset-ps)
	}

	LinkerPageTmpl.Handle(w, r, data)
}
