package privateforum

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

// pendingActionTTL allows a bounded two-hour stale-form reauthentication window.
const pendingActionTTL = 2 * time.Hour

// StartGroupDiscussionPage renders a dedicated page to start a private group discussion.
func StartGroupDiscussionPage(w http.ResponseWriter, r *http.Request) {
	if err := renderStartGroupDiscussionPage(w, r, &forumhandlers.CreateTopicPageForm{}); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("prepare private topic form: %w", err))
	}
}

func renderStartGroupDiscussionPage(w http.ResponseWriter, r *http.Request, formData *forumhandlers.CreateTopicPageForm) error {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Start private group discussion"
	if _, err := cd.DeleteExpiredPendingActions(r.Context()); err != nil {
		return fmt.Errorf("clean expired pending actions: %w", err)
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("generate form nonce: %w", err)
	}
	formNonce := hex.EncodeToString(b)
	formNonceHash := sha256.Sum256([]byte(formNonce))
	browserID := core.GetBrowserID(w, r)
	if browserID == "" {
		return fmt.Errorf("generate browser binding")
	}
	now := time.Now().UTC()
	if err := cd.CreatePendingAction(r.Context(), db.SystemInsertPendingActionParams{
		ID:         hex.EncodeToString(formNonceHash[:]),
		Uid:        cd.UserID,
		BrowserID:  browserID,
		ActionType: string(TaskPrivateTopicCreate),
		FormData:   "",
		CreatedAt:  now,
		ExpiresAt:  now.Add(pendingActionTTL),
	}); err != nil {
		return fmt.Errorf("insert pending action form nonce: %w", err)
	}

	data := struct {
		CreateTask tasks.TaskString
		FormData   *forumhandlers.CreateTopicPageForm
		FormNonce  string
	}{CreateTask: TaskPrivateTopicCreate, FormData: formData, FormNonce: formNonce}
	if err := PrivateForumStartDiscussionPageTmpl.Handle(w, r, data); err != nil {
		return fmt.Errorf("render private topic form: %w", err)
	}
	return nil
}

const PrivateForumStartDiscussionPageTmpl tasks.Template = "domains/privateforum/start_discussion.gohtml"
