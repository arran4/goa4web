package privateforum

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/internal/tasks"
)

// StartGroupDiscussionPage renders a dedicated page to start a private group discussion.
func StartGroupDiscussionPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	// Page title/header as requested
	cd.PageTitle = "Start private group discussion"

	b := make([]byte, 32)
	var formNonce string
	if _, err := rand.Read(b); err == nil {
		formNonce = hex.EncodeToString(b)
		formNonceHash := sha256.Sum256([]byte(formNonce))
		browserID := core.GetBrowserID(w, r)
		if err := cd.Queries().InsertPendingAction(r.Context(), db.InsertPendingActionParams{
			ID:         hex.EncodeToString(formNonceHash[:]),
			Uid:        cd.UserID,
			BrowserID:  browserID,
			ActionType: string(TaskPrivateTopicCreate),
			FormData:   "",
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}); err != nil {
			log.Printf("insert pending action nonce fail: %v", err)
			handlers.RenderErrorPage(w, r, errors.New("internal error"))
			return
		}
	}

	data := struct {
		CreateTask tasks.TaskString
		FormData   *forumhandlers.CreateTopicPageForm
		FormNonce  string
	}{CreateTask: TaskPrivateTopicCreate, FormData: &forumhandlers.CreateTopicPageForm{}, FormNonce: formNonce}
	_ = PrivateForumStartDiscussionPageTmpl.Handle(w, r, data)
}

const PrivateForumStartDiscussionPageTmpl tasks.Template = "domains/privateforum/start_discussion.gohtml"
