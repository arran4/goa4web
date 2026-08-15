package forum

import (
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/gorilla/mux"
)

func AdminThreadsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Threads"

	_ = ForumAdminThreadsPageTmpl.Handle(w, r, struct{}{})
}

const ForumAdminThreadsPageTmpl tasks.Template = "domains/forum/adminThreadsPage.gohtml"

func AdminThreadDeletePage(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(mux.Vars(r)["thread"])
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	topicID, err := cd.Queries().GetForumTopicIdByThreadId(r.Context(), int32(threadID))
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	if err := ThreadDelete(r.Context(), cd.Queries(), int32(threadID), topicID); err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	http.Redirect(w, r, "/admin/forum/threads", http.StatusSeeOther)
}

func AdminThreadDeleteConfirmPage(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(mux.Vars(r)["thread"])
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "/admin/forum/threads", err)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Confirm forum thread delete"
	data := struct {
		Message      string
		ConfirmLabel string
		Back         string
	}{
		Message:      "Are you sure you want to delete forum thread " + strconv.Itoa(threadID) + "?",
		ConfirmLabel: "Confirm delete",
		Back:         "/admin/forum/thread/" + strconv.Itoa(threadID),
	}
	_ = ConfirmPageTmpl.Handle(w, r, data)
}

const ConfirmPageTmpl tasks.Template = "pages/misc/confirmPage.gohtml"

func AdminThreadPage(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(mux.Vars(r)["thread"])
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "/admin/forum/threads", err)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	threadRow, err := cd.Queries().AdminGetForumThreadById(r.Context(), int32(threadID))
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "/admin/forum/threads", err)
		return
	}

	cd.PageTitle = "Forum Admin Thread"
	data := struct {
		Thread *db.AdminGetForumThreadByIdRow
	}{
		Thread: threadRow,
	}

	_ = ForumAdminThreadPageTmpl.Handle(w, r, data)
}

const ForumAdminThreadPageTmpl tasks.Template = "domains/forum/adminThreadPage.gohtml"
