package forum

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type DraftsTask struct{ tasks.TaskString }

var (
	draftsTask = &DraftsTask{TaskString: "draft"}
)

func (dt *DraftsTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)

	uid, _ := session.Values["UID"].(int32)

	switch r.Method {
	case http.MethodGet:
		return dt.get(w, r, cd, uid)
	case http.MethodPost:
		return dt.post(w, r, cd, uid)
	case http.MethodDelete:
		return dt.delete(w, r, cd, uid)
	default:
		return handlers.ErrBadRequest
	}
}

func (dt *DraftsTask) get(w http.ResponseWriter, r *http.Request, cd *common.CoreData, uid int32) any {
	threadID := cd.SelectedThreadID()

	draftIDVal := r.URL.Query().Get("id")
	if draftIDVal != "" {
		draftID, err := strconv.Atoi(draftIDVal)
		if err != nil {
			return fmt.Errorf("parsing draft id: %w", err)
		}
		draft, err := cd.GetDraft(r.Context(), int32(draftID), uid)
		if err != nil {
			return fmt.Errorf("getting draft: %w", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(draft)
		return nil
	}

	drafts, err := cd.ListDraftsForThread(r.Context(), int32(threadID), uid)
	if err != nil {
		return fmt.Errorf("listing drafts: %w", err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drafts)
	return nil
}

func (dt *DraftsTask) post(w http.ResponseWriter, r *http.Request, cd *common.CoreData, uid int32) any {
	text := r.PostFormValue("replytext")
	draftIDVal := r.PostFormValue("draft_id")
	draftName := r.PostFormValue("draft_name")

	threadID := cd.SelectedThreadID()

	draftID, _ := strconv.Atoi(draftIDVal)

	if draftName == "" {
		draftName = "Draft from " + time.Now().Format("2006-01-02 15:04:05")
	}

	var newDraftID int64
	var err error
	if draftID > 0 {
		err = cd.UpdateDraft(r.Context(), db.UpdateDraftParams{
			ID:      int32(draftID),
			UserID:  uid,
			Name:    draftName,
			Content: text,
		})
		if err != nil {
			return fmt.Errorf("updating draft: %w", err)
		}
		newDraftID = int64(draftID)
	} else {
		newDraftID, err = cd.CreateDraft(r.Context(), db.CreateDraftParams{
			UserID:   uid,
			ThreadID: threadID,
			Name:     draftName,
			Content:  text,
		})
		if err != nil {
			return fmt.Errorf("creating draft: %w", err)
		}
	}

	err = cd.Queries().AddContentPrivateLabel(r.Context(), db.AddContentPrivateLabelParams{
		Item:   "thread",
		ItemID: threadID,
		UserID: uid,
		Label:  "has draft",
	})
	if err != nil {
		return fmt.Errorf("adding private label: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"draft_id": newDraftID,
	})
	return nil
}

func (dt *DraftsTask) delete(w http.ResponseWriter, r *http.Request, cd *common.CoreData, uid int32) any {
	draftIDVal := r.URL.Query().Get("id")
	draftID, err := strconv.Atoi(draftIDVal)
	if err != nil {
		return fmt.Errorf("parsing draft id: %w", err)
	}

	err = cd.Queries().DeleteDraft(r.Context(), db.DeleteDraftParams{
		ID:     int32(draftID),
		UserID: uid,
	})
	if err != nil {
		return fmt.Errorf("deleting draft: %w", err)
	}

	threadID := cd.SelectedThreadID()

	hasDrafts, err := cd.HasDrafts(r.Context(), threadID, uid)
	if err != nil {
		return fmt.Errorf("checking for drafts: %w", err)
	}

	if !hasDrafts {
		err = cd.Queries().RemoveContentPrivateLabel(r.Context(), db.RemoveContentPrivateLabelParams{
			Item:   "thread",
			ItemID: int32(threadID),
			UserID: uid,
			Label:  "has draft",
		})
		if err != nil {
			return fmt.Errorf("removing private label: %w", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})
	return nil
}
