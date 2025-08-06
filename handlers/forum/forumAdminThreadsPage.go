package forum

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
)

func AdminThreadsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Threads"

	handlers.TemplateHandler(w, r, "adminThreadsPage.gohtml", struct{}{})
}

func AdminThreadDeletePage(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(mux.Vars(r)["thread"])
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	topicID, err := cd.Queries().GetForumTopicIdByThreadId(r.Context(), int32(threadID))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if err := ThreadDelete(r.Context(), cd.Queries(), int32(threadID), topicID); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/admin/forum/threads", http.StatusTemporaryRedirect)
}

func AdminThreadDeleteConfirmPage(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(mux.Vars(r)["thread"])
	if err != nil {
		http.Redirect(w, r, "/admin/forum/threads?error="+err.Error(), http.StatusTemporaryRedirect)
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
	handlers.TemplateHandler(w, r, "confirmPage.gohtml", data)
}

func AdminThreadPage(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(mux.Vars(r)["thread"])
	if err != nil {
		http.Redirect(w, r, "/admin/forum/threads?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	session, _ := core.GetSession(r)
	var uid int32
	if session != nil {
		uid, _ = session.Values["UID"].(int32)
	}

	threadRow, err := cd.Queries().GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      int32(threadID),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		http.Redirect(w, r, "/admin/forum/threads?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	cd.PageTitle = "Forum Admin Thread"
	data := struct {
		Thread *db.GetThreadLastPosterAndPermsRow
	}{
		Thread: threadRow,
	}

	handlers.TemplateHandler(w, r, "adminThreadPage.gohtml", data)
}
